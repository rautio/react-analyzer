use crate::ts_config::{get_aliases, get_base_path, get_closest};

use super::languages::ParsedFile;
use super::languages::TestFile;
use super::package_json::{list_dependencies, PackageJson};
use super::ts_config::TypeScriptConfig;
use serde::Serialize;
use std::cmp::Ordering;
use std::collections::HashMap;
use std::path::{Component, Path, PathBuf};

#[derive(Serialize)]
pub struct Summary {
    pub line_count: usize,
    pub import_count: usize,
    pub file_count: usize,
    pub unused_file_count: usize,
}
impl std::fmt::Display for Summary {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        write!(
            f,
            "Total Files:     {}\nTotal Lines:     {}\nTotal Imports:   {}\nDead Files:      {}",
            self.file_count, self.line_count, self.import_count, self.unused_file_count
        )
    }
}

#[derive(Serialize)]
pub struct Output {
    pub import_graph: ImportGraph,
    pub dead_files: Vec<String>,
    pub unknown_imports: Vec<String>,
    pub exports: Vec<FileExports>,
    pub summary: Summary,
    pub package_json: PackageJsonExtract,
}

#[derive(Serialize)]
pub struct ImportGraph {
    pub nodes: Vec<Node>,
    pub edges: Vec<Edge>,
}

#[derive(Clone, Debug, Serialize, PartialEq, Eq, PartialOrd)]
pub struct Edge {
    pub id: usize,
    pub source: usize,
    pub target: usize,
    pub is_default: bool,
    pub name: String,
}

impl Ord for Edge {
    fn cmp(&self, other: &Self) -> Ordering {
        self.id.cmp(&other.id)
    }
}

#[derive(Clone, Debug, Serialize, PartialEq, Eq, PartialOrd)]
pub struct Node {
    pub id: usize,
    pub path: String,
    pub file_name: Option<String>,
    pub extension: Option<String>,
    pub line_count: Option<usize>,
}

impl Ord for Node {
    fn cmp(&self, other: &Self) -> Ordering {
        self.id.cmp(&other.id)
    }
}
pub fn normalize_path(path: &Path) -> PathBuf {
    let mut components = path.components().peekable();
    let mut ret = if let Some(c @ Component::Prefix(..)) = components.peek().cloned() {
        components.next();
        PathBuf::from(c.as_os_str())
    } else {
        PathBuf::new()
    };

    for component in components {
        match component {
            Component::Prefix(..) => unreachable!(),
            Component::RootDir => {
                ret.push(component.as_os_str());
            }
            Component::CurDir => {}
            Component::ParentDir => {
                ret.pop();
            }
            Component::Normal(c) => {
                ret.push(c);
            }
        }
    }
    ret
}

pub fn extract_dead_files(
    graph: &ImportGraph,
    dependencies: Vec<String>,
    root: &Path,
) -> (Vec<String>, Vec<String>) {
    let mut connected_nodes: HashMap<usize, bool> = HashMap::new();
    // Iterate edges to gather all nodes that are imported or references
    for e in &graph.edges {
        connected_nodes.insert(e.source, true);
        connected_nodes.insert(e.target, true);
    }
    let mut dead_files: Vec<String> = Vec::new();
    let mut unknown_imports: Vec<String> = Vec::new();
    for n in &graph.nodes {
        if !connected_nodes.contains_key(&n.id) {
            // Check if the path is a dependency, if so skip
            let mut src = PathBuf::from("");
            let mut is_dep = false;
            for c in PathBuf::from(&n.path).components() {
                src.push(c);
                if dependencies.contains(&src.display().to_string()) {
                    is_dep = true;
                }
            }
            if !is_dep {
                let ext = match n.clone().extension {
                    Some(extension) => extension,
                    None => String::from(""),
                };
                let file_path = Path::new(root).join(Path::new(&n.path)).with_extension(ext);
                if file_path.exists() {
                    dead_files.push(n.path.clone());
                } else {
                    unknown_imports.push(n.path.clone());
                }
            }
        }
    }
    return (dead_files, unknown_imports);
}

pub fn extract_import_graph(
    files: &Vec<ParsedFile>,
    ts_configs: &Vec<TypeScriptConfig>,
) -> ImportGraph {
    let mut node_count = 0;
    let mut edge_count = 0;
    let mut node_map: HashMap<String, Node> = HashMap::new();
    let mut edges: Vec<Edge> = Vec::new();
    for file in files {
        let ts_config = get_closest(ts_configs, PathBuf::from(&file.path));
        let aliases = get_aliases(ts_config.cloned());
        let base_path = get_base_path(ts_config.cloned());
        let file_path = &file.path;
        let path = PathBuf::from(&file.path).with_extension("");
        let file_name = match path.file_name() {
            Some(n) => match n.to_str() {
                Some(s) => s,
                None => "",
            },
            None => "",
        };
        let dir = match &path.parent() {
            Some(path) => path.clone().display().to_string(),
            None => String::from(""),
        };
        // Way may have mapped an "index" file in which case only the directory name exists in the node_map
        if node_map.contains_key(&dir) && file_name == "index" {
            // Mapping to the parent
            let old = node_map.get(&dir).unwrap();
            let id = old.id;
            let line_count = old.line_count;
            let real = PathBuf::from(&file.path);
            node_map.remove(&dir);
            node_map.insert(
                file_path.to_string(),
                Node {
                    id,
                    path: file_path.to_string(),
                    file_name: Some(real.file_name().unwrap().to_str().unwrap().to_string()),
                    extension: Some(real.extension().unwrap().to_str().unwrap().to_string()),
                    line_count,
                },
            );
        }
        // Create current file node
        if !node_map.contains_key(file_path) {
            node_map.insert(
                file_path.to_string(),
                Node {
                    id: node_count,
                    path: file_path.to_string(),
                    file_name: Some(file.name.clone()),
                    extension: Some(file.extension.clone()),
                    line_count: Some(file.line_count),
                },
            );
            node_count += 1;
        } else {
            // Exists, make sure we have all data populated
            let mut node = node_map.get_mut(file_path).unwrap();
            if node.file_name == None {
                node.file_name = Some(file.name.clone());
            }
            if node.extension == None {
                node.extension = Some(file.extension.clone());
            }
            if node.line_count == None {
                node.line_count = Some(file.line_count);
            }
        }
        // Create source file nodes and edges
        for import in &file.imports {
            // Normalize import path to a real path or npm module
            let mut src = import.source.clone();
            // Could be: NPM module, alias or genuinely a relative import.
            if src.starts_with(".") {
                // Genuine relative path
                let mut file_path = PathBuf::from(&import.file_path);
                // Get to the directory
                file_path.pop();
                let source_path = Path::new(&file_path).join(Path::new(&src));
                // Normalize to a real path
                src = normalize_path(&source_path).display().to_string();
            } else {
                match aliases.clone() {
                    Some(aliases) => {
                        let ts_config_path = &ts_config.unwrap().file_path;
                        for alias in aliases.clone().into_keys() {
                            let mut my_alias = alias.as_str();
                            let mut value = aliases.get(&alias).unwrap()[0].as_str();
                            // Ends with '*' means it matches on subpaths.
                            if my_alias.ends_with(r"*") {
                                my_alias = my_alias.strip_suffix(r"*").unwrap();
                                value = value.strip_suffix(r"*").unwrap();
                            }
                            if src.starts_with(&my_alias) {
                                // Aliases are relative to the ts_config location
                                let mut path = PathBuf::from(ts_config_path.clone().unwrap());
                                path.pop(); // Last component is the file name
                                            // Need to account for a base path if one is specified
                                match base_path.clone() {
                                    Some(base_path) => {
                                        path = path.join(PathBuf::from(base_path));
                                    }
                                    None => {}
                                }
                                let replaced = src.replace(&my_alias, value);
                                path = path.join(PathBuf::from(&replaced));
                                // Normalize the final path
                                src = normalize_path(&PathBuf::from(path)).display().to_string();
                            }
                        }
                    }
                    None => {}
                }
            }
            if src.ends_with('/') {
                src.pop();
            }
            if !node_map.contains_key(&src) {
                node_map.insert(
                    src.to_string(),
                    Node {
                        id: node_count,
                        path: src.to_string(),
                        file_name: None,
                        extension: None,
                        line_count: None,
                    },
                );
                node_count += 1;
            }
            // Map all named imports to this source
            for name in &import.named {
                edges.push(Edge {
                    id: edge_count,
                    source: node_map.get(&src).unwrap().id,
                    target: node_map.get(file_path).unwrap().id,
                    is_default: import.is_default,
                    name: name.to_string(),
                });
                edge_count += 1;
            }
            // If it has a default, map that as well
            if import.is_default {
                edges.push(Edge {
                    id: edge_count,
                    source: node_map.get(&src).unwrap().id,
                    target: node_map.get(file_path).unwrap().id,
                    is_default: import.is_default,
                    name: String::from(""),
                });
                edge_count += 1;
            }
        }
    }
    let nodes = node_map.values().cloned().collect::<Vec<Node>>();
    return ImportGraph { nodes, edges };
}

#[derive(Serialize)]
pub struct Export {
    name: String,
    target: String,
    is_default: bool,
}

#[derive(Serialize)]
pub struct FileExports {
    source: String,
    exports: Vec<Export>,
}

#[derive(Debug)]
struct Target {
    id: usize,
    name: String,
    is_default: bool,
}

pub fn extract_exports(import_graph: &ImportGraph) -> Vec<FileExports> {
    let mut file_exports: Vec<FileExports> = Vec::new();
    // Local mapping for easier lookup
    let mut node_id_map: HashMap<usize, &Node> = HashMap::new();
    for node in &import_graph.nodes {
        node_id_map.insert(node.id, node);
    }
    // Maps: source_id -> [target_id1, target_id2]
    // source_id is the file exporting. target ids are the files importing
    let mut export_map: HashMap<usize, Vec<Target>> = HashMap::new();
    for edge in &import_graph.edges {
        if export_map.contains_key(&edge.source) {
            export_map.get_mut(&edge.source).unwrap().push(Target {
                id: edge.target,
                name: edge.name.clone(),
                is_default: edge.is_default,
            });
        } else {
            export_map.insert(
                edge.source,
                vec![Target {
                    id: edge.target,
                    name: edge.name.clone(),
                    is_default: edge.is_default,
                }],
            );
        }
    }
    // Iterate and parse all known exports
    for source in export_map.keys() {
        let targets = export_map.get(source).unwrap();
        let source_file = node_id_map.get(source).unwrap();
        let mut exports = Vec::new();
        for target in targets {
            exports.push(Export {
                name: target.name.clone(),
                target: node_id_map.get(&target.id).unwrap().path.clone(),
                is_default: target.is_default,
            })
        }
        file_exports.push(FileExports {
            source: source_file.path.clone(),
            exports,
        })
    }
    return file_exports;
}

#[derive(Serialize, Debug, Clone)]
pub struct PackageJsonExtract {
    dependencies: HashMap<String, usize>, // Does not account for monorepo
}

pub fn extract_package_json(
    files: &Vec<ParsedFile>,
    package_jsons: Vec<PackageJson>,
) -> PackageJsonExtract {
    let mut dependencies: HashMap<String, usize> = HashMap::new();
    for p_json in package_jsons {
        for d in list_dependencies(p_json) {
            dependencies.insert(d, 0);
        }
    }
    for f in files {
        for import in f.imports.iter() {
            let splits = import.source.split('/');
            let mut package = String::from("");
            if import.source.starts_with(".") {
                // Package can't sort with "." - it must be a file import
                continue;
            }
            for s in splits {
                package.push_str(s);
                if dependencies.contains_key(&package) {
                    *dependencies.get_mut(&package).unwrap() += 1;
                    break;
                }
                package.push_str(r"/");
            }
        }
    }
    return PackageJsonExtract { dependencies };
}

pub fn extract(
    root: &Path,
    files: Vec<ParsedFile>,
    package_jsons: Vec<PackageJson>,
    ts_configs: Vec<TypeScriptConfig>,
) -> Output {
    let file_count = files.len();
    let mut line_count = 0;
    let mut import_count: usize = 0;
    let import_graph = extract_import_graph(&files, &ts_configs);
    let package_json = extract_package_json(&files, package_jsons);
    let dependencies = package_json.clone().dependencies.into_keys().collect();
    let (dead_files, unknown_imports) = extract_dead_files(&import_graph, dependencies, root);
    let exports = extract_exports(&import_graph);
    for file in files {
        line_count += file.line_count;
        import_count += file.imports.len();
    }
    let summary = Summary {
        line_count,
        import_count,
        file_count,
        unused_file_count: dead_files.len(),
    };
    return Output {
        import_graph,
        dead_files,
        unknown_imports,
        exports,
        summary,
        package_json,
    };
}

#[derive(Serialize)]
pub struct TestOutput {}
pub struct TestSummary {
    count: usize,
    skipped_count: usize,
    line_count: usize,
}

impl std::fmt::Display for TestSummary {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        write!(
            f,
            "Total Tests:     {}\nSkipped Tests:   {}\nTotal Lines:     {}",
            self.count, self.skipped_count, self.line_count
        )
    }
}

pub fn extract_test_files(test_files: Vec<TestFile>) -> (TestSummary, TestOutput) {
    let mut test_count = 0;
    let mut skipped_test_count = 0;
    let mut test_line_count = 0;
    for test_file in &test_files {
        test_count += test_file.test_count;
        skipped_test_count += test_file.skipped_test_count;
        test_line_count += test_file.line_count;
    }
    return (
        TestSummary {
            count: test_count,
            skipped_count: skipped_test_count,
            line_count: test_line_count,
        },
        TestOutput {},
    );
}

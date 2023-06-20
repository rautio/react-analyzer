use super::parser::ParsedFile;
use serde::Serialize;
use std::collections::HashMap;

pub struct Summary {
    pub line_count: usize,
    pub import_count: usize,
    pub file_count: usize,
}
impl std::fmt::Display for Summary {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        write!(
            f,
            "Total Files:     {}\nTotal Lines:     {}\nTotal Imports:   {}",
            self.file_count, self.line_count, self.import_count
        )
    }
}

#[derive(Serialize)]
pub struct Output {
    import_graph: ImportGraph,
}

#[derive(Serialize)]
pub struct ImportGraph {
    nodes: Vec<Node>,
    edges: Vec<Edge>,
}

#[derive(Serialize)]
pub struct Edge {
    id: usize,
    source: usize,
    target: usize,
}

#[derive(Clone, Debug, Serialize)]
pub struct Node {
    id: usize,
    path: String,
    file_name: Option<String>,
    extension: Option<String>,
    line_count: Option<usize>,
}

pub fn extract_import_graph(files: &Vec<ParsedFile>) -> ImportGraph {
    let mut node_count = 0;
    let mut edge_count = 0;
    let mut node_map: HashMap<String, Node> = HashMap::new();
    let mut edges: Vec<Edge> = Vec::new();
    for file in files {
        let file_path = &file.path;
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
            let mut src = import.source.clone();
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
            edges.push(Edge {
                id: edge_count,
                source: node_map.get(&src).unwrap().id,
                target: node_map.get(file_path).unwrap().id,
            });
            edge_count += 1;
        }
    }
    let nodes = node_map.values().cloned().collect();
    return ImportGraph { nodes, edges };
}

pub fn extract(files: Vec<ParsedFile>) -> (Summary, Output) {
    let file_count = files.len();
    let mut line_count = 0;
    let mut import_count: usize = 0;
    let import_graph = extract_import_graph(&files);
    for file in files {
        line_count += file.line_count;
        import_count += file.imports.len();
    }
    return (
        Summary {
            line_count,
            import_count,
            file_count,
        },
        Output { import_graph },
    );
}

use super::Language;
use crate::languages::{Export, Import, ParsedFile, TestFile};
use lazy_static::lazy_static;
use regex::Regex;
use rome_js_parser;
use rome_rowan::AstNode;
use std::fs;
use std::fs::File;
use std::io::BufRead;
use std::io::BufReader;
use std::io::Error;
use std::path::{Path, PathBuf};

lazy_static! {
    static ref TEST_REGEX: Regex = Regex::new(r#"(test|it)\(('|").*('|"),"#,).unwrap();
    static ref SKIPPED_REGEX: Regex = Regex::new(r#"(test.skip|it.skip)\(('|").*('|"),"#,).unwrap();
}

pub struct JavaScript {}

impl JavaScript {
    /// JS file name given a path. If a file name is Component/index.js the name becomes Component.
    pub fn get_file_name(&self, path: &Path) -> String {
        let mut name = path.file_stem().unwrap();
        // If its an index file we want to use the folder as the file name
        if name == "index" {
            name = &path.parent().unwrap().file_stem().unwrap();
        }
        return name.to_str().unwrap().to_string();
    }
    pub fn parse_module(
        &self,
        file_string: &String,
        file_path: &String,
    ) -> (Vec<Import>, Vec<Export>) {
        let mut imports: Vec<Import> = Vec::new();
        let mut exports: Vec<Export> = Vec::new();
        let parsed = rome_js_parser::parse_module(file_string);
        for item in parsed.tree().items() {
            if item.as_js_import().is_some() {
                // Import
                if item.as_js_import().unwrap().import_clause().is_err() {
                    continue;
                }
                let import_clause = item.as_js_import().unwrap().import_clause().unwrap();
                if import_clause.as_js_import_named_clause().is_some() {
                    // Named import!
                    let named_clause = import_clause.as_js_import_named_clause().unwrap();
                    let named_imports = named_clause.as_fields().named_import.unwrap();
                    if named_imports.as_js_named_import_specifiers().is_some() {
                        let import_specifiers = named_imports
                            .as_js_named_import_specifiers()
                            .unwrap()
                            .as_fields()
                            .specifiers;
                        let mut source = String::from("");
                        if named_clause.as_fields().source.is_ok() {
                            source = named_clause
                                .as_fields()
                                .source
                                .unwrap()
                                .text()
                                .replace(&['\'', '"'][..], "");
                        }
                        imports.push(Import {
                            file_path: file_path.clone(),
                            source,
                            is_default: named_clause.as_fields().default_specifier.is_some(),
                            named: import_specifiers
                                .syntax()
                                .to_string()
                                .split(',')
                                .map(str::trim)
                                .map(str::to_string)
                                .collect::<Vec<String>>(),
                            line: 0,
                        });
                    }
                }
                if import_clause.as_js_import_default_clause().is_some() {
                    // Default import!
                    let default = import_clause.as_js_import_default_clause().unwrap();
                    let mut source = String::from("");
                    if default.as_fields().source.is_ok() {
                        source = default
                            .as_fields()
                            .source
                            .unwrap()
                            .text()
                            .replace(&['\'', '"'][..], "");
                    }
                    imports.push(Import {
                        file_path: file_path.clone(),
                        source,
                        is_default: true,
                        named: Vec::new(),
                        line: 0,
                    });
                }
                if import_clause.as_js_import_namespace_clause().is_some() {
                    let default = import_clause.as_js_import_namespace_clause().unwrap();
                    let mut source = String::from("");
                    if default.as_fields().source.is_ok() {
                        source = default
                            .as_fields()
                            .source
                            .unwrap()
                            .text()
                            .replace(&['\'', '"'][..], "");
                    }
                    imports.push(Import {
                        file_path: file_path.clone(),
                        source,
                        is_default: true,
                        named: Vec::new(),
                        line: 0,
                    });
                }
            }
            if item.as_js_export().is_some() {
                let export = item.as_js_export().unwrap();
                if export.export_clause().is_err() {
                    continue;
                }
                let export_clause = export.export_clause().unwrap();

                if export_clause.as_js_export_named_from_clause().is_some() {
                    // Export named from
                    exports.push(Export {
                        file_path: file_path.clone(),
                        line: 0,
                        named: export_clause
                            .as_js_export_named_from_clause()
                            .unwrap()
                            .as_fields()
                            .specifiers
                            .syntax()
                            .to_string()
                            .split(',')
                            .map(str::trim)
                            .map(str::to_string)
                            .collect::<Vec<String>>(),
                        default: String::from(""),
                        source: export_clause
                            .as_js_export_named_from_clause()
                            .unwrap()
                            .as_fields()
                            .source
                            .unwrap()
                            .to_string(),
                    })
                }
                if export_clause.as_js_export_named_clause().is_some() {
                    // Named exports
                    exports.push(Export {
                        file_path: file_path.clone(),
                        line: 0,
                        named: export_clause
                            .as_js_export_named_clause()
                            .unwrap()
                            .as_fields()
                            .specifiers
                            .syntax()
                            .to_string()
                            .split(',')
                            .map(str::trim)
                            .map(str::to_string)
                            .collect::<Vec<String>>(),
                        default: String::from(""),
                        source: String::from(""),
                    })
                }
                if export_clause.as_js_export_from_clause().is_some() {
                    let mut source = String::from("");
                    if export_clause
                        .as_js_export_from_clause()
                        .unwrap()
                        .as_fields()
                        .source
                        .is_ok()
                    {
                        source = export_clause
                            .as_js_export_from_clause()
                            .unwrap()
                            .as_fields()
                            .source
                            .unwrap()
                            .to_string();
                    }
                    // Export from
                    exports.push(Export {
                        file_path: file_path.clone(),
                        line: 0,
                        default: String::from(""),
                        named: vec![],
                        source,
                    })
                }
                if export_clause
                    .as_js_export_default_expression_clause()
                    .is_some()
                {
                    let mut default = String::from("");
                    if export_clause
                        .as_js_export_default_expression_clause()
                        .unwrap()
                        .as_fields()
                        .expression
                        .is_ok()
                    {
                        default = export_clause
                            .as_js_export_default_expression_clause()
                            .unwrap()
                            .as_fields()
                            .expression
                            .unwrap()
                            .to_string();
                    }
                    exports.push(Export {
                        file_path: file_path.clone(),
                        line: 0,
                        named: vec![],
                        default,
                        source: String::from(""),
                    })
                }
            }
        }
        return (imports, exports);
    }
}

impl Language for JavaScript {
    fn parse_file(&self, path: &Path, root_prefix: PathBuf) -> Result<ParsedFile, Error> {
        let file_string = fs::read_to_string(&path).expect(&format!(
            "Unable to read file: {}",
            &path.display().to_string()
        ));
        let file_path = path
            .strip_prefix(root_prefix)
            .unwrap()
            .display()
            .to_string();
        let (imports, exports) = self.parse_module(&file_string, &file_path);
        let parsed = ParsedFile {
            line_count: file_string.lines().count(),
            imports,
            exports,
            name: self.get_file_name(&path),
            extension: path.extension().unwrap().to_str().unwrap().to_string(),
            path: file_path,
        };
        return Ok(parsed);
    }
    fn parse_test_file(&self, path: &Path) -> Result<TestFile, Error> {
        let file = File::open(path)?;
        let reader = BufReader::new(file);
        let mut line_count = 0;
        let mut test_count = 0;
        let mut skipped_test_count = 0;
        for l in reader.lines() {
            if let Ok(cur_line) = l {
                if let Some(_) = TEST_REGEX.find(&cur_line) {
                    test_count += 1;
                }
                if let Some(_) = SKIPPED_REGEX.find(&cur_line) {
                    skipped_test_count += 1;
                }
            }
            line_count += 1;
        }
        let parsed: TestFile = TestFile {
            name: self.get_file_name(&path),
            path: path.to_str().unwrap().to_string(),
            line_count,
            test_count,
            skipped_test_count,
        };
        return Ok(parsed);
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn test_file_name() -> Result<(), String> {
        let js = JavaScript {};
        // Most file names
        assert_eq!(
            js.get_file_name(&Path::new("/MyTest/component.tsx")),
            String::from("component")
        );
        assert_eq!(
            js.get_file_name(&Path::new("/MyTest/test.spec.js")),
            String::from("test.spec")
        );
        // index file names
        assert_eq!(
            js.get_file_name(&Path::new("/MyComponent/index.jsx")),
            String::from("MyComponent")
        );
        Ok(())
    }
}

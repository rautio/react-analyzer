use super::Language;
use crate::languages::{Export, Import, ParsedFile, TestFile};
use lazy_static::lazy_static;
use regex::Regex;
use rome_js_parser;
use rome_js_syntax;
use rome_js_syntax::JsSyntaxKind;
use std::fs;
use std::fs::File;
use std::io::BufRead;
use std::io::BufReader;
use std::io::Error;
use std::path::Path;

lazy_static! {
    static ref IMPORT_REGEX: Regex = Regex::new(
        r#"^import\s+?((?:(?:(?:[\w*\s{},]*)\s)+from\s+?)|)(?:(?:"(.*?)")|(?:'(.*?)'))[\s]*?(?:;|$|)"#,
    )
    .unwrap();
    static ref IMPORT_NAMES_REGEX: Regex = Regex::new(r"\s?(.*?),?(\{(.*)\})? from?").unwrap();
    static ref TEST_REGEX: Regex = Regex::new(r#"(test|it)\(('|").*('|"),"#,).unwrap();
    static ref SKIPPED_REGEX: Regex = Regex::new(r#"(test.skip|it.skip)\(('|").*('|"),"#,).unwrap();
    static ref VARIABLE_REGEX: Regex = Regex::new(r"^\s?(let|var|const)\s?(.*) =").unwrap();
    static ref EXPORT_REGEX: Regex = Regex::new(r"^export (.+)").unwrap();
}

pub struct JavaScript {}

impl JavaScript {
    pub fn get_file_name(&self, path: &Path) -> String {
        let mut name = path.file_stem().unwrap();
        // If its an index file we want to use the folder as the file name
        if name == "index" {
            name = &path.parent().unwrap().file_stem().unwrap();
        }
        return name.to_str().unwrap().to_string();
    }
    pub fn parse_module(&self, path: &Path) -> (Vec<Import>, Vec<Export>) {
        let mut imports: Vec<Import> = Vec::new();
        let mut exports: Vec<Export> = Vec::new();
        let file_string = fs::read_to_string(&path).expect("Unable to read file");
        let parsed = rome_js_parser::parse_module(&file_string);
        let parsed_imports = parsed
            .syntax()
            .descendants()
            .filter(|node| node.kind() == JsSyntaxKind::JS_IMPORT);
        let parsed_exports = parsed
            .syntax()
            .descendants()
            .filter(|node| node.kind() == JsSyntaxKind::JS_EXPORT);
        for import in parsed_imports {
            let mut source = String::from("");
            let mut is_default = false;
            let mut named: Vec<String> = Vec::new();
            for im in import.descendants() {
                if im.kind() == JsSyntaxKind::JS_MODULE_SOURCE {
                    source = im.to_string();
                }
                if im.kind() == JsSyntaxKind::JS_IMPORT_DEFAULT_CLAUSE {
                    is_default = true;
                }
                if im.kind() == JsSyntaxKind::JS_NAMED_IMPORT_SPECIFIER_LIST {
                    named = im
                        .to_string()
                        .split(',')
                        .map(str::trim)
                        .map(str::to_string)
                        .collect();
                }
            }
            imports.push(Import {
                source: source,
                is_default,
                default: String::from(""), // TODO: remove
                named,
                line: 0,
            })
        }
        for export in parsed_exports {
            for im in export.descendants() {
                let mut default = String::from("");
                let named = Vec::new();
                if im.kind() == JsSyntaxKind::JS_EXPORT_DEFAULT_EXPRESSION_CLAUSE {
                    default = im.to_string();
                }
                exports.push(Export {
                    file_path: path.display().to_string(),
                    line: 0,
                    named,
                    default,
                    source: String::from(""),
                })
            }
        }
        return (imports, exports);
    }
}

impl Language for JavaScript {
    fn parse_file(&self, path: &Path) -> Result<ParsedFile, Error> {
        let file = File::open(path)?;
        let reader = BufReader::new(file);
        let (imports, exports) = self.parse_module(&path);
        let mut line_count = 0;
        let mut variable_count = 0;
        for l in reader.lines() {
            if let Ok(cur_line) = l {
                if VARIABLE_REGEX.is_match(&cur_line) {
                    variable_count += 1;
                }
            }
            line_count += 1;
        }
        let parsed = ParsedFile {
            line_count,
            imports,
            exports,
            name: self.get_file_name(&path),
            extension: path.extension().unwrap().to_str().unwrap().to_string(),
            path: path.to_str().unwrap().to_string(),
            variable_count,
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

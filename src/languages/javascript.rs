use super::Language;
use crate::languages;
use crate::languages::Export;
use crate::languages::ParsedFile;
use crate::languages::TestFile;
use lazy_static::lazy_static;
use path_absolutize::*;
use regex::Regex;
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
    static ref EXPORT_REGEX: Regex = Regex::new(r"^export (.*)").unwrap();
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
    /// Is the given string an import statement.
    pub fn is_import(&self, line: &String) -> bool {
        if let Some(_) = IMPORT_REGEX.find(&line) {
            return true;
        }
        return false;
    }
    /// Is the given string an export statement
    pub fn is_export(&self, line: &String) -> bool {
        if let Some(_) = EXPORT_REGEX.find(&line) {
            return true;
        }
        return false;
    }
    pub fn parse_import(&self, line: &String, current_path: &Path) -> languages::Import {
        if !IMPORT_REGEX.is_match(&line) {
            panic!("Not an import statement");
        }
        let captures = IMPORT_REGEX.captures(&line).unwrap();
        // Capture imported names
        let raw_import_names = captures.get(1).map_or("", |m| m.as_str());
        // Then import could not have any named imports like: "import 'style.css';"
        let mut default_import = "";
        let mut named_imports = "";
        let name_captures_option = IMPORT_NAMES_REGEX.captures(raw_import_names);
        if !name_captures_option.is_none() {
            let name_captures = name_captures_option.unwrap();
            // let name_captures = IMPORT_NAMES_REGEX.captures(raw_import_names).unwrap();
            default_import = name_captures.get(1).map_or("", |m| m.as_str());
            named_imports = name_captures.get(3).map_or("", |m| m.as_str());
        }
        // Capture import source path
        let double_quote_import = captures.get(2).map_or("", |m| m.as_str());
        let mut source = double_quote_import;
        if source == "" {
            let single_quote_import = captures.get(3).map_or("", |m| m.as_str());
            source = single_quote_import;
        }
        if source == "." {
            source = "";
        }
        let mut named = Vec::new();
        if !named_imports.is_empty() {
            named = named_imports
                .split(',')
                .map(str::trim)
                .map(str::to_string)
                .collect();
        }
        let source_path = Path::new(&source);
        // Relative path, convert it to an absolute path
        if source_path.to_str().unwrap().to_string().starts_with('.') {
            return languages::Import {
                source: current_path
                    // current_path includes filename
                    .parent()
                    .unwrap()
                    // join with relative import
                    .join(source_path)
                    .absolutize()
                    .unwrap()
                    .file_stem()
                    .unwrap()
                    .to_str()
                    .unwrap()
                    .to_string(),
                named,
                default: default_import.to_string(),
            };
        }
        // Either an alias, absolute path or node_module
        return languages::Import {
            source: source.to_string(),
            named,
            default: default_import.to_string(),
        };
    }
    pub fn parse_export(&self, line: &String, current_path: &Path) -> languages::Export {
        let captures: regex::Captures<'_> = EXPORT_REGEX.captures(&line).unwrap();
        let default_export = "";
        let named_exports = captures.get(1).map_or("", |m| m.as_str());
        return Export {
            file_path: current_path.display().to_string(),
            named: named_exports.split(',').map(str::to_string).collect(),
            default: default_export.to_string(),
            source: String::from(""),
        };
    }
}

impl Language for JavaScript {
    fn parse_file(&self, path: &Path) -> Result<ParsedFile, Error> {
        let file = File::open(path)?;
        let reader = BufReader::new(file);
        let mut imports = Vec::new();
        let mut exports: Vec<Export> = Vec::new();
        let mut line_count = 0;
        let mut variable_count = 0;
        for l in reader.lines() {
            if let Ok(cur_line) = l {
                if self.is_import(&cur_line) {
                    imports.push(self.parse_import(&cur_line, &path))
                }
                if VARIABLE_REGEX.is_match(&cur_line) {
                    variable_count += 1;
                }
                if self.is_export(&cur_line) {
                    exports.push(self.parse_export(&cur_line, &path))
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

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn test_is_import() {
        let lang = JavaScript {};
        let true_values = [
            "import videos from './videos/index.js'",
            "import * from \"foo\"",
            "import test, { bar } from \"foo\"",
            "import rick from \"morty\"",
            "import { rick, roll } from \"foo\";",
            "import { rick, roll } from 'foo';",
            "import * from 'foo';",
            "import 'module-name'",
            "import \"module-name\"",
            "import {
                something
            } from \"./test/okbb\"",
        ];
        for val in true_values {
            assert_eq!(lang.is_import(&String::from(val)), true);
        }
        let false_values = [
            "import* from 'foo';",
            "import * from \"foo';",
            "const f = 2;",
            "import \"module-name'",
            "importing hya from 'ttt'",
            "import fbsfrom ''",
        ];
        for val in false_values {
            assert_eq!(lang.is_import(&String::from(val)), false);
        }
    }
}

use super::Language;
use crate::languages;
use lazy_static::lazy_static;
use path_absolutize::*;
use regex::Regex;
use std::path::Path;

lazy_static! {
    static ref IMPORT_REGEX: Regex = Regex::new(
        r#"^import(\s+?(?:(?:(?:[\w*\s{},]*)\s)+from\s+?)|)(?:(?:"(.*?)")|(?:'(.*?)'))[\s]*?(?:;|$|)"#,
    )
    .unwrap();
}

pub struct JavaScript {}

impl Language for JavaScript {
    fn get_file_name(&self, path: &Path) -> String {
        let mut name = path.file_stem().unwrap();
        // If its an index file we want to use the folder as the file name
        if name == "index" {
            name = &path.parent().unwrap().file_stem().unwrap();
        }
        return name.to_str().unwrap().to_string();
    }
    fn is_import(&self, line: &String) -> bool {
        if let Some(_) = IMPORT_REGEX.find(&line) {
            return true;
        }
        return false;
    }
    fn parse_import(&self, line: &String, current_path: &Path) -> languages::Import {
        let captures = IMPORT_REGEX.captures(&line).unwrap();
        let names = captures.get(1).map_or("", |m| m.as_str());
        let double_quote_import = captures.get(2).map_or("", |m| m.as_str());
        let mut source = double_quote_import;
        if source == "" {
            let single_quote_import = captures.get(3).map_or("", |m| m.as_str());
            source = single_quote_import;
        }
        if source == "." {
            source = "";
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
                names: vec![names.to_string()],
            };
        }
        // Either an alias, absolute path or node_module
        return languages::Import {
            source: source.to_string(),
            names: vec![names.to_string()],
        };
    }
}

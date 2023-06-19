use super::Language;
use crate::languages;
use lazy_static::lazy_static;
use regex::Regex;

lazy_static! {
    static ref IMPORT_REGEX: Regex = Regex::new(
        r#"^import(\s+?(?:(?:(?:[\w*\s{},]*)\s)+from\s+?)|)(?:((?:".*?")|(?:'.*?')))[\s]*?(?:;|$|)"#,
    )
    .unwrap();
}

pub struct JavaScript {}

impl Language for JavaScript {
    fn is_import(&self, line: &String) -> bool {
        if let Some(_) = IMPORT_REGEX.find(&line) {
            return true;
        }
        return false;
    }
    fn parse_import(&self, line: &String) -> languages::Import {
        let captures = IMPORT_REGEX.captures(&line).unwrap();
        let names = captures.get(1).map_or("", |m| m.as_str());
        let source = captures.get(2).map_or("", |m| m.as_str());
        return languages::Import {
            source: source.to_string(),
            names: vec![names.to_string()],
        };
    }
}

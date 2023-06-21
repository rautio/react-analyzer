use super::languages::javascript::JavaScript;
use super::languages::Import;
use super::languages::Language;
use lazy_static::lazy_static;
use regex::Regex;
use std::fs::File;
use std::io::BufRead;
use std::io::BufReader;
use std::io::Error;
use std::path::Path;

lazy_static! {
    static ref TEST_REGEX: Regex = Regex::new(r#"(test|it)\(('|").*('|"),"#,).unwrap();
    static ref SKIPPED_REGEX: Regex = Regex::new(r#"(test.skip|it.skip)\(('|").*('|"),"#,).unwrap();
}

#[derive(Clone, Debug)]
pub struct ParsedFile {
    pub line_count: usize,
    pub imports: Vec<Import>,
    pub name: String,
    pub extension: String,
    pub path: String,
}

pub struct TestFile {
    pub line_count: usize,
    pub name: String,
    pub path: String,
    pub test_count: usize,
    pub skipped_test_count: usize,
}

const JS: JavaScript = JavaScript {};

fn get_language(path: &Path) -> (JavaScript, Option<&str>) {
    let lang = match path.extension() {
        None => JS,
        Some(os_str) => match os_str.to_str() {
            Some("js") => JS,
            Some("ts") => JS,
            Some("jsx") => JS,
            Some("tsx") => JS,
            _ => panic!("You forgot to specify this case!"),
        },
    };
    return (lang, path.extension().unwrap().to_str());
}

impl ParsedFile {
    pub fn new(path: &Path) -> Result<Self, Error> {
        let file = File::open(path)?;
        let reader = BufReader::new(file);
        let (lang, extension) = get_language(path);
        let mut imports = Vec::new();
        let mut line_count = 0;
        for l in reader.lines() {
            if let Ok(cur_line) = l {
                if lang.is_import(&cur_line) {
                    imports.push(lang.parse_import(&cur_line, &path))
                }
            }
            line_count += 1;
        }
        let parsed = ParsedFile {
            line_count,
            imports,
            name: lang.get_file_name(&path),
            extension: extension.unwrap().to_string(),
            path: path.to_str().unwrap().to_string(),
        };
        return Ok(parsed);
    }
}

impl TestFile {
    pub fn new(path: &Path) -> Result<Self, Error> {
        let file = File::open(path)?;
        let reader = BufReader::new(file);
        let lang = JavaScript {};
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
            name: lang.get_file_name(&path),
            path: path.to_str().unwrap().to_string(),
            line_count,
            test_count,
            skipped_test_count,
        };
        return Ok(parsed);
    }
}

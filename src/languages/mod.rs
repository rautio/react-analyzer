pub mod javascript;
pub mod typescript;
pub mod unknown;
use std::io::Error;
use std::path::Path;

use self::javascript::JavaScript;
use self::typescript::TypeScript;
use self::unknown::Unknown;

#[derive(Clone, Debug)]

pub struct Import {
    pub source: String,
    pub names: Vec<String>,
}
#[derive(Clone, Debug)]
pub struct ParsedFile {
    pub line_count: usize,
    pub imports: Vec<Import>,
    pub name: String,
    pub extension: String,
    pub path: String,
    pub variable_count: usize,
}

pub struct TestFile {
    pub line_count: usize,
    pub name: String,
    pub path: String,
    pub test_count: usize,
    pub skipped_test_count: usize,
}

pub trait Language {
    fn parse_file(&self, path: &Path) -> Result<ParsedFile, Error>;
    fn parse_test_file(&self, path: &Path) -> Result<TestFile, Error>;
}

const JS: JavaScript = JavaScript {};
const TS: TypeScript = TypeScript {};
const UK: Unknown = Unknown {};

// Need a way to dynamically get language struct from file extension
pub fn parse_file(path: &Path) -> Result<ParsedFile, Error> {
    return match path.extension() {
        None => UK.parse_file(&path), // Default for any unknown extensions
        Some(os_str) => match os_str.to_str() {
            Some("js") => JS.parse_file(&path),
            Some("ts") => TS.parse_file(&path),
            Some("jsx") => JS.parse_file(&path),
            Some("tsx") => TS.parse_file(&path),
            _ => panic!("You forgot to specify this case!"),
        },
    };
}
pub fn parse_test_file(path: &Path) -> Result<TestFile, Error> {
    return match path.extension() {
        None => panic!("Unrecognized file extension"),
        Some(os_str) => match os_str.to_str() {
            Some("js") => JS.parse_test_file(&path),
            Some("ts") => TS.parse_test_file(&path),
            Some("jsx") => JS.parse_test_file(&path),
            Some("tsx") => TS.parse_test_file(&path),
            _ => panic!("You forgot to specify this case!"),
        },
    };
}

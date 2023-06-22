pub mod javascript;
use std::io::Error;
use std::path::Path;

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

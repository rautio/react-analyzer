use super::Language;
use crate::languages::ParsedFile;
use crate::languages::TestFile;
use std::fs::File;
use std::io::BufRead;
use std::io::BufReader;
use std::io::Error;
use std::path::Path;

pub struct Unknown {}

impl Language for Unknown {
    fn parse_file(&self, path: &Path) -> Result<ParsedFile, Error> {
        let file = File::open(path)?;
        let reader = BufReader::new(file);
        let parsed = ParsedFile {
            line_count: reader.lines().count(),
            imports: Vec::new(),
            exports: Vec::new(),
            name: path.file_name().unwrap().to_str().unwrap().to_string(),
            extension: path.extension().unwrap().to_str().unwrap().to_string(),
            path: path.to_str().unwrap().to_string(),
            variable_count: 0,
        };
        return Ok(parsed);
    }
    fn parse_test_file(&self, path: &Path) -> Result<TestFile, Error> {
        let file = File::open(path)?;
        let reader = BufReader::new(file);
        let parsed: TestFile = TestFile {
            name: path.file_name().unwrap().to_str().unwrap().to_string(),
            path: path.to_str().unwrap().to_string(),
            line_count: reader.lines().count(),
            test_count: 0,
            skipped_test_count: 0,
        };
        return Ok(parsed);
    }
}

use super::javascript::JavaScript;
use super::Language;
use crate::languages::ParsedFile;
use crate::languages::TestFile;
use std::io::Error;
use std::path::{Path, PathBuf};

pub struct TypeScript {}

const JS: JavaScript = JavaScript {};

impl TypeScript {}

impl Language for TypeScript {
    fn parse_file(&self, path: &Path, prefix: PathBuf) -> Result<ParsedFile, Error> {
        return JS.parse_file(path, prefix);
    }
    fn parse_test_file(&self, path: &Path) -> Result<TestFile, Error> {
        return JS.parse_test_file(path);
    }
}

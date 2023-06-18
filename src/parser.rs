use std::fs::File;
use std::io::BufRead;
use std::io::BufReader;
use std::io::Error;
use std::path::Path;

#[derive(Clone, Debug)]
pub struct ParsedFile {
    pub line_count: usize,
}

impl ParsedFile {
    pub fn new(path: &Path) -> Result<Self, Error> {
        let file = File::open(path)?;
        let reader = BufReader::new(file);
        let parsed = ParsedFile {
            line_count: reader.lines().count(),
        };
        return Ok(parsed);
    }
}

// import\s+?(?:(?:(?:[\w*\s{},]*)\s+from\s+?)|)(?:(?:".*?")|(?:'.*?'))[\s]*?(?:;|$|)

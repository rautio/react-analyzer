use std::io::BufRead;
use std::path::Path;
use std::fs::File;
use std::io::BufReader;
use std::io::Error;

#[derive(Clone, Debug)]
pub struct ParsedFile {
    pub line_count: usize,
}

// impl ParsedFile {
//     fn new<R: BufRead>(mut reader: R) -> Result<Self, Error> {
//         let mut parsed = ParsedFile {
//             line_count: 0
//         };
//         Ok(parsed)
//         Error()
//     }
// }

pub fn parse_file(path:&Path) -> Result<ParsedFile, Error> {
    let file = File::open(path)?;
    let reader = BufReader::new(file);
    let parsed = ParsedFile{
        line_count: reader.lines().count()
    };
    return Ok(parsed)
}
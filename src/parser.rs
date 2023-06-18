use regex::Regex;
use std::ffi::OsStr;
use std::fs::File;
use std::io::BufRead;
use std::io::BufReader;
use std::io::Error;
use std::path::Path;

#[derive(Clone, Debug)]
pub struct Import {
    file: String,
    line: usize,
}
#[derive(Clone, Debug)]
pub struct ParsedFile {
    pub line_count: usize,
    pub imports: Vec<Import>,
    // pub extension: String,
}

// fn get_extension_from_path(path: &Path) -> Option<&str> {
//     path.extension().and_then(OsStr::to_str)
// }

impl ParsedFile {
    pub fn new(path: &Path) -> Result<Self, Error> {
        let file = File::open(path)?;
        let reader = BufReader::new(file);
        let mut imports = Vec::new();
        let mut line_count = 0;
        let import_re: Regex = Regex::new(r#"^import\s+?(?:(?:(?:[\w*\s{},]*)\s+from\s+?)|)(?:(?:".*?")|(?:'.*?'))[\s]*?(?:;|$|)"#).unwrap();
        for l in reader.lines() {
            if let Ok(cur_line) = l {
                if let Some(_) = import_re.find(&cur_line) {
                    let statement = import_re.find(&cur_line).unwrap().as_str();
                    // println!("{}", statement);
                    imports.push(Import {
                        line: line_count,
                        file: statement.to_string(),
                    })
                }
            }
            line_count += 1;
        }
        // let mut extension = get_extension_from_path(path);
        let parsed = ParsedFile {
            line_count,
            imports,
            // extension,
        };
        return Ok(parsed);
    }
}

use super::languages::javascript::JavaScript;
use super::languages::Import;
use super::languages::Language;
use std::fs::File;
use std::io::BufRead;
use std::io::BufReader;
use std::io::Error;
use std::path::Path;

#[derive(Clone, Debug)]
pub struct ParsedFile {
    pub line_count: usize,
    pub imports: Vec<Import>,
    pub name: String,
    pub extension: String,
    pub path: String,
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
            name: path.file_name().unwrap().to_str().unwrap().to_string(),
            extension: extension.unwrap().to_string(),
            path: path.to_str().unwrap().to_string(),
        };
        return Ok(parsed);
    }
}

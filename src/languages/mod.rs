pub mod javascript;
use std::path::Path;

#[derive(Clone, Debug)]

pub struct Import {
    pub source: String,
    pub names: Vec<String>,
}

pub trait Language {
    fn is_import(&self, line: &String) -> bool;
    fn parse_import(&self, line: &String, current_path: &Path) -> Import;
}

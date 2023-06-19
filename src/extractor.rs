use super::parser::ParsedFile;

pub struct Summary {
    pub line_count: usize,
    pub import_count: usize,
    pub file_count: usize,
}
impl std::fmt::Display for Summary {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        write!(
            f,
            "Total Files:     {}\nTotal Lines:     {}\nTotal Imports:   {}",
            self.file_count, self.line_count, self.import_count
        )
    }
}
pub struct Output {}
pub struct ImportTree {}

pub fn extract_import_tree(files: Vec<ParsedFile>) -> ImportTree {
    return ImportTree {};
}

pub fn extract(files: Vec<ParsedFile>) -> (Summary, Output) {
    let file_count = files.len();
    let mut line_count = 0;
    let mut import_count: usize = 0;
    for file in files {
        line_count += file.line_count;
        import_count += file.imports.len();
    }
    return (
        Summary {
            line_count,
            import_count,
            file_count,
        },
        Output {},
    );
}

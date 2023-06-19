use super::parser::ParsedFile;
use regex::Regex;
use std::fs::metadata;
use std::path::Path;

/// Lists all files in given diretory path.
fn list_files(path: &Path, pattern: &Regex, ignore_pattern: &Regex) -> Vec<ParsedFile> {
    let mut files: Vec<ParsedFile> = Vec::new();
    // Read path and validate
    for entry in path.read_dir().expect("Unable to read directory.") {
        if let Ok(entry) = entry {
            let md = metadata(entry.path()).unwrap();
            // If matches ignore, skip
            let name = &entry.path().display().to_string();
            if ignore_pattern.is_match(name) {
                continue;
            }
            if md.is_dir() {
                files.append(&mut list_files(&entry.path(), pattern, ignore_pattern));
            } else {
                // Only add file if it matches pattern
                if pattern.is_match(name) {
                    let parsed = ParsedFile::new(&entry.path());
                    if let Ok(p) = parsed {
                        files.push(p)
                    }
                }
            }
        }
    }
    return files;
}

/// Scan a given path
pub fn scan(path: &Path) {
    println!("Scanning: {}", path.display());

    // Add as CLI parameters and read from ignore file
    let ignore_pattern: Regex = Regex::new(r"node_modules").unwrap();
    let pattern = Regex::new(r"^.*\.(jsx|js|tsx|ts)$").unwrap();

    let files = list_files(path, &pattern, &ignore_pattern);
    println!("Files: {}", files.len());
    let mut total_lines = 0;
    let mut total_imports: usize = 0;
    for file in files {
        total_lines += file.line_count;
        total_imports += file.imports.len();
    }
    println!("Total Lines: {}", total_lines);
    println!("Total Imports: {}", total_imports);
}

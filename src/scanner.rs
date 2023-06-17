use std::path::Path;
use std::path::PathBuf;
use std::fs::metadata;
use regex::Regex;

/// Lists all files in given diretory path. 
fn list_files(path: &Path, pattern: &Regex, ignore_pattern: &Regex) -> Vec<PathBuf> {
    let mut files : Vec<PathBuf> = Vec::new();
    // Read path and validate
    for entry in path.read_dir().expect("Unable to read directory.") {
        if let Ok(entry) = entry {
            let md = metadata(entry.path()).unwrap();
            // If matches ignore, skip
            let name= &entry.path().display().to_string();
            if ignore_pattern.is_match(name) {
                continue
            }
            if md.is_dir() {
                files.append(&mut list_files(&entry.path(), pattern, ignore_pattern));
            } else {
                // Only add file if it matches pattern
                if pattern.is_match(name) {
                    files.push(entry.path())
                }
            }
        }
    }
    return files;

}

/// Scan a given path
pub fn scan(path:&Path) {
    println!("Scanning: {}", path.display());

    // Add as CLI parameters and read from ignore file
    let ignore_pattern: Regex = Regex::new(r"node_modules").unwrap();
    let pattern = Regex::new(r"^.*\.jsx|js|tsx|ts$").unwrap();

    let files = list_files(path, &pattern, &ignore_pattern);
    println!("Files: {}", files.len())
}
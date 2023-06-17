use std::path::Path;
use std::path::PathBuf;
use std::fs::metadata;

fn list_files(path: &Path) -> Vec<PathBuf> {
    let mut files : Vec<PathBuf> = Vec::new();
    // Read path and validate
    for entry in path.read_dir().expect("Unable to read directory.") {
        if let Ok(entry) = entry {
            let md = metadata(entry.path()).unwrap();
            if md.is_dir() {
                files.append(&mut list_files(&entry.path()))
            } else {
                files.push(entry.path())
            }
        }
    }
    return files;

}

pub fn scan(path:&Path) {
    println!("Scanning: {}", path.display());
    let files = list_files(path);
    println!("Files: {}", files.len())
}
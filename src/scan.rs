use crate::languages::parse_file;
use crate::languages::parse_test_file;
use crate::languages::ParsedFile;
use crate::languages::TestFile;
use regex::Regex;
use std::fs::metadata;
use std::thread;
use std::sync::mpsc::channel;
use std::path::Path;
use std::time::Instant;

fn find_files(root_path: &Path, pattern: &Regex, ignore_pattern: &Regex) -> Vec<String> {
    let mut files: Vec<String> = Vec::new();
    // Read path and validate
    for entry in root_path.read_dir().expect("Unable to read directory.") {
        if let Ok(entry) = entry {
            let file_path = &entry.path();
            let md = metadata(file_path).unwrap();
            // If matches ignore, skip
            let name = file_path.display().to_string();
            if ignore_pattern.is_match(&name) {
                continue;
            }
            if md.is_dir() {
                files.append(&mut find_files(&file_path, pattern, ignore_pattern));
            } else {
                // Only add file if it matches pattern
                if pattern.is_match(&name) {
                    files.push(file_path.to_str().unwrap().to_string());
                }
            }
        }
    }
    return files;
}
/// Scan a given path and return all files parsed
pub fn scan(root_path: &Path, pattern: &Regex, ignore_pattern: &Regex) -> Vec<ParsedFile> {
    let now = Instant::now();
    let files: Vec<String> = find_files(root_path, pattern, ignore_pattern);
    let mut parsed_files: Vec<ParsedFile> = Vec::new();
    let (tx, rx) = channel();
    let threads: Vec<_> = files.into_iter().map(|path|  {
        let tx = tx.clone();
        thread::spawn(move || {
            let file_path = Path::new(&path);
            let parsed = parse_file(&file_path);
            if let Ok(p) = parsed {
                tx.send(p).unwrap();
            }
        })
    }).collect();
    for handle in threads {
        handle.join().unwrap();
        let p = rx.recv().unwrap();
        parsed_files.push(p);
    }
    let elapsed = now.elapsed();
    println!("Scan done in: {:.2?}!", elapsed);
    return parsed_files;
}

pub fn scan_test_files(root_path: &Path, pattern: &Regex, ignore_pattern: &Regex) -> Vec<TestFile> {
    let files: Vec<String> = find_files(root_path, pattern, ignore_pattern);
    let mut test_files: Vec<TestFile> = Vec::new();
    for path in files {
        let parsed = parse_test_file(Path::new(&path));
        if let Ok(t) = parsed {
            test_files.push(t);
        }
    }
    return test_files;
}

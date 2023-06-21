use crate::languages::javascript::JavaScript;
use crate::languages::Language;
use crate::languages::ParsedFile;
use crate::languages::TestFile;
use regex::Regex;
use std::fs::metadata;
use std::path::Path;

const JS: JavaScript = JavaScript {};

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
    let files: Vec<String> = find_files(root_path, pattern, ignore_pattern);
    let mut parsed_files: Vec<ParsedFile> = Vec::new();
    for path in files {
        let parsed = JS.parse_file(Path::new(&path));
        if let Ok(p) = parsed {
            parsed_files.push(p);
        }
    }
    return parsed_files;
}

pub fn scan_test_files(root_path: &Path, pattern: &Regex, ignore_pattern: &Regex) -> Vec<TestFile> {
    let files: Vec<String> = find_files(root_path, pattern, ignore_pattern);
    let mut test_files: Vec<TestFile> = Vec::new();
    for path in files {
        let parsed = JS.parse_test_file(Path::new(&path));
        if let Ok(t) = parsed {
            test_files.push(t);
        }
    }
    let mut test_count = 0;
    let mut skipped_test_count = 0;
    let mut test_line_count = 0;
    for test_file in &test_files {
        test_count += test_file.test_count;
        skipped_test_count += test_file.skipped_test_count;
        test_line_count += test_file.line_count;
    }
    println!("");
    println!("Test Count: {}", test_count);
    println!("Skipped Test Count: {}", skipped_test_count);
    println!("Test Line Count: {}", test_line_count);
    return test_files;
}

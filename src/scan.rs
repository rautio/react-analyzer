use crate::languages::parse_file;
use crate::languages::parse_test_file;
use crate::languages::ParsedFile;
use crate::languages::TestFile;
use crate::package_json;
use crate::package_json::PackageJson;
use regex::Regex;
use std::fs;
use std::fs::metadata;
use std::path::{Path, PathBuf};
use std::sync::mpsc::channel;
use std::time::Instant;
use threadpool::ThreadPool;

struct Files {
    all_files: Vec<String>,
    package_json: Vec<PathBuf>,
    ts_config: Vec<PathBuf>,
}

fn find_files(root_path: &Path, pattern: &Regex, ignore_pattern: &Regex) -> Files {
    let mut all_files: Vec<String> = Vec::new();
    let mut package_json: Vec<PathBuf> = Vec::new();
    let mut ts_config: Vec<PathBuf> = Vec::new();
    // Read path and validate
    for entry in root_path.read_dir().expect("Unable to read directory.") {
        if let Ok(entry) = entry {
            let file_path = &entry.path();
            if file_path.file_name().unwrap() == "package.json" {
                package_json.push(file_path.to_path_buf());
            }
            if file_path.file_name().unwrap() == "tsconfig.json" {
                ts_config.push(file_path.to_path_buf());
            }
            let md = metadata(file_path).unwrap();
            // If matches ignore, skip
            let name = file_path.display().to_string();
            if ignore_pattern.is_match(&name) {
                continue;
            }
            if md.is_dir() {
                let f = &mut find_files(&file_path, pattern, ignore_pattern);
                all_files.append(&mut f.all_files);
                package_json.append(&mut f.package_json);
                ts_config.append(&mut f.ts_config);
            } else {
                // Only add file if it matches pattern
                if pattern.is_match(&name) {
                    all_files.push(file_path.to_str().unwrap().to_string());
                }
            }
        }
    }
    return Files {
        all_files,
        package_json,
        ts_config,
    };
}
/// Scan a given path and return all files parsed
pub fn scan(
    root_path: &Path,
    pattern: &Regex,
    ignore_pattern: &Regex,
) -> (Vec<ParsedFile>, Vec<PackageJson>) {
    let now = Instant::now();
    let f = find_files(root_path, pattern, ignore_pattern);
    let mut parsed_files: Vec<ParsedFile> = Vec::new();
    let parsed_package_jsons: Vec<PackageJson> = package_json::parse(f.package_json);
    // We need to configure a fixed number of workers so we don't hit OS limits. On Mac the
    // max number of open files is 256 and this can easily be hit if running in a large repo.
    let n_workers = 64; // The performance bottleneck becomes file I/O and not number of threads after a certain point
    let pool = ThreadPool::new(n_workers);
    let (tx, rx) = channel();
    for t_json in f.ts_config {
        let file_string = fs::read_to_string(&t_json).expect(&format!(
            "Unable to read file: {}",
            &t_json.display().to_string()
        ));
        let _: serde_json::Value = serde_json::from_str(file_string.as_str()).expect(&format!(
            "JSON was not well-formatted in: {}",
            &t_json.display().to_string()
        ));
    }
    let threads: Vec<_> = f
        .all_files
        .into_iter()
        .map(|path| {
            let tx = tx.clone();
            let prefix = PathBuf::from(root_path);
            pool.execute(move || {
                let file_path = Path::new(&path);
                let parsed = parse_file(&file_path, prefix);
                if let Ok(p) = parsed {
                    tx.send(p).unwrap();
                }
            })
        })
        .collect();
    for _ in threads {
        let p = rx.recv().unwrap();
        parsed_files.push(p);
    }
    let elapsed = now.elapsed();
    println!("Scan done in: {:.2?}!", elapsed);
    return (parsed_files, parsed_package_jsons);
}

pub fn scan_test_files(root_path: &Path, pattern: &Regex, ignore_pattern: &Regex) -> Vec<TestFile> {
    let f = find_files(root_path, pattern, ignore_pattern);
    let mut test_files: Vec<TestFile> = Vec::new();
    for path in f.all_files {
        let parsed = parse_test_file(Path::new(&path));
        if let Ok(t) = parsed {
            test_files.push(t);
        }
    }
    return test_files;
}

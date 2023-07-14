use clap::Parser;
use regex::Regex;
use std::path::Path;
use std::time::Instant;
mod extract;
mod languages;
mod output;
mod package_json;
mod path_utils;
mod print;
mod scan;
mod ts_config;

#[derive(Parser)]
#[command(author, version)]
#[command(about = "react-analyzer - a static code analyzer for React based projects")]
struct Cli {
    /// Path to folder root
    path: std::path::PathBuf,
    // language: String,
}

fn main() {
    let now = Instant::now();
    {
        print::welcome_message();
        // Parse command line arguments
        let args = Cli::parse();
        let root = Path::new(&args.path);
        // Default patterns. Need cli or config file to override.
        let pattern = Regex::new(r"^.*\.(jsx|js|tsx|ts)$").unwrap();
        let ignore_pattern: Regex = Regex::new(r".*.test.js").unwrap();
        let test_pattern: Regex = Regex::new(r".*.(cy|test|spec|unit).(jsx|tsx|js|ts)$").unwrap();
        print::input(
            root,
            pattern.clone(),
            ignore_pattern.clone(),
            test_pattern.clone(),
        );
        // Scan Files
        let (files, package_jsons, ts_configs) = scan::scan(root, &pattern, &ignore_pattern);
        let output = extract::extract(files, package_jsons, ts_configs);
        let _ = output::write_output(&output);
        println!("=== File Summary ===\n{}\n", output.summary);
        // Scan Test Files
        let test_files: Vec<languages::TestFile> =
            scan::scan_test_files(root, &test_pattern, &ignore_pattern);
        let (test_summary, _) = extract::extract_test_files(test_files);
        println!("=== Test Summary ===\n{}\n", test_summary);
    }
    let elapsed = now.elapsed();
    println!("Done in: {:.2?}!", elapsed);
}

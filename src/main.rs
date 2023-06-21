use clap::Parser;
use regex::Regex;
use std::path::Path;
use std::time::Instant;
mod extractor;
mod languages;
mod output;
mod scanner;

#[derive(Parser)]
#[command(author, version)]
#[command(about = "norppa - a static code analyzer for React based projects")]
struct Cli {
    /// Path to folder root
    path: std::path::PathBuf,
    // language: String,
}

fn main() {
    let now = Instant::now();
    {
        // Parse command line arguments
        let args = Cli::parse();
        let root = Path::new(&args.path);
        println!("Analyzing: {}", root.display());
        // TODO: Add as CLI parameters and read from ignore file
        let pattern = Regex::new(r"^.*\.(jsx|js|tsx|ts)$").unwrap();
        println!("Scan pattern: {}", pattern);
        let ignore_pattern: Regex = Regex::new(r"node_modules|.*.test.js").unwrap();
        println!("Ignore pattern: {}", ignore_pattern);
        let files: Vec<languages::ParsedFile> = scanner::scan(root, &pattern, &ignore_pattern);
        let test_pattern: Regex = Regex::new(r".*.(cy|test|spec|unit).(jsx|tsx|js|ts)").unwrap();
        let _: Vec<languages::TestFile> =
            scanner::scan_test_files(root, &test_pattern, &ignore_pattern);
        let (summary, output) = extractor::extract(files);
        let _ = output::write_output(output);
        println!("\n{}\n", summary);
    }
    let elapsed = now.elapsed();
    println!("Done in: {:.2?}!", elapsed);
}

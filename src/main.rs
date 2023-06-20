use clap::Parser;
use std::path::Path;
use std::time::Instant;
mod extractor;
mod languages;
mod output;
mod parser;
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
        let files: Vec<parser::ParsedFile> = scanner::scan(root);
        let (summary, output) = extractor::extract(files);
        let _ = output::write_output(output);
        println!("\n{}\n", summary);
    }
    let elapsed = now.elapsed();
    println!("Done in: {:.2?}!", elapsed);
}

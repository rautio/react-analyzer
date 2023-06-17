use clap::Parser;
use std::path::Path;
use std::time::Instant;
mod scanner;
mod parser;

#[derive(Parser)]
#[command(author, version)]
#[command(about = "norppa - a static code analyzer for React based projects")]
struct Cli {
    /// Path to folder root
    path: std::path::PathBuf,
}

fn main() {
    let now = Instant::now();
    {
        // Parse command line arguments
        let args = Cli::parse();
        let path = Path::new(&args.path);
        scanner::scan(path);
    }
    let elapsed = now.elapsed();
    println!("Time: {:.2?}", elapsed);
}

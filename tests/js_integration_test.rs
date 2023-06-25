use norppa::languages::javascript::JavaScript;
use std::fs::File;
use std::io::BufRead;
use std::io::BufReader;
use std::path::PathBuf;

fn read_file(file_path: &str) -> Result<BufReader<File>, Box<dyn std::error::Error>> {
    let mut d = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    d.push(file_path);
    let file = File::open(d)?;
    return Ok(BufReader::new(file));
}

#[test]
fn test_is_import() -> Result<(), Box<dyn std::error::Error>> {
    let lang = JavaScript {};
    let reader = read_file("tests/mocks/import.js").unwrap();
    for line in reader.lines() {
        // Some formats aren't yet supported.
        let l = &line?;
        if !l.starts_with("//") {
            assert_eq!(lang.is_import(l), true);
        }
    }
    Ok(())
}

#[test]
fn test_is_export() -> Result<(), Box<dyn std::error::Error>> {
    let lang = JavaScript {};
    let reader = read_file("tests/mocks/export.js").unwrap();
    for line in reader.lines() {
        // Some formats aren't yet supported.
        let l = &line?;
        if !l.starts_with("//") && !l.is_empty() {
            assert_eq!(lang.is_export(l), true);
        }
    }
    Ok(())
}

use norppa::languages::javascript::JavaScript;
use std::fs::File;
use std::io::BufRead;
use std::io::BufReader;
use std::path::PathBuf;

#[test]
fn test_is_import_file() -> Result<(), Box<dyn std::error::Error>> {
    let lang = JavaScript {};
    let mut d = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    d.push("tests/mocks/import.js");
    let file = File::open(d)?;
    let reader = BufReader::new(file);
    for line in reader.lines() {
        // Some formats aren't yet supported.
        let l = &line?;
        if !l.starts_with("//") {
            assert_eq!(lang.is_import(l), true);
        }
    }
    Ok(())
}

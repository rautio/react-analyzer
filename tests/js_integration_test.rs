use react_analyzer::languages::javascript::JavaScript;
use react_analyzer::languages::Import;
use std::fs::File;
use std::io::BufRead;
use std::io::BufReader;
use std::path::Path;
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

#[test]
fn test_parse_import() -> Result<(), Box<dyn std::error::Error>> {
    let lang = JavaScript {};
    let reader = read_file("tests/mocks/import.js").unwrap();
    let expected = [
        Import {
            source: String::from("module-name"),
            default: String::from("defaultExport"),
            named: Vec::new(),
        },
        Import {
            source: String::from("module-name"),
            default: String::from("* as name"),
            named: Vec::new(),
        }, // Wrong for now
        Import {
            source: String::from("module-name"),
            default: String::from(""),
            named: [String::from("export1")].to_vec(),
        },
        Import {
            source: String::from("module-name"),
            default: String::from(""),
            named: [String::from("export1 as alias1")].to_vec(),
        }, // Wrong for now
        Import {
            source: String::from("module-name"),
            default: String::from(""),
            named: [String::from("default as alias")].to_vec(),
        }, // Wrong for now
        Import {
            source: String::from("module-name"),
            default: String::from(""),
            named: [String::from("export1"), String::from("export2")].to_vec(),
        },
        Import {
            source: String::from("module-name"),
            default: String::from(""),
            named: Vec::new(),
        }, // skipped
        Import {
            source: String::from("module-name"),
            default: String::from(""),
            named: Vec::new(),
        }, // skipped
        Import {
            source: String::from("module-name"),
            default: String::from(""),
            named: Vec::new(),
        }, // skipped
        Import {
            source: String::from("module-name"),
            default: String::from("defaultExport, * as name"),
            named: Vec::new(),
        }, // Wrong for now
        Import {
            source: String::from("module-name"),
            default: String::from(""),
            named: Vec::new(),
        },
    ];
    for (i, line) in reader.lines().enumerate() {
        let l = &line?;
        if !l.starts_with("//") && !l.is_empty() {
            let import = lang.parse_import(l, Path::new("/"));
            let expect = &expected[i];
            assert_eq!(import.source, expect.source);
            assert_eq!(import.named, expect.named);
            assert_eq!(import.default, expect.default);
        }
    }
    Ok(())
}

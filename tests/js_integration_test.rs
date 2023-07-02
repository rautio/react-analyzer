use react_analyzer::languages::javascript::JavaScript;
use react_analyzer::languages::Import;
use std::fs;
use std::path::PathBuf;

#[test]
fn test_parse_module() -> Result<(), Box<dyn std::error::Error>> {
    let lang = JavaScript {};
    let mut path = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    path.push("tests/mocks/import.js");
    let file_string = fs::read_to_string(&path).expect("Unable to read file");
    let (imports, _) = lang.parse_module(&file_string, String::from(""));
    let expected = [
        Import {
            source: String::from("\"module-name\""),
            named: Vec::new(),
            is_default: true,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            named: Vec::new(),
            is_default: true,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            named: [String::from("export1")].to_vec(),
            is_default: false,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            named: [String::from("export1 as alias1")].to_vec(), // Wrong for now
            is_default: false,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            named: [String::from("default as alias")].to_vec(), // Wrong for now
            is_default: false,                                  // Wrong
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            named: [String::from("export1"), String::from("export2")].to_vec(),
            is_default: false,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            named: [
                String::from("export1"),
                String::from("export2 as alias2"),
                String::from("/* … */"),
            ]
            .to_vec(),
            is_default: false,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            named: [String::from("\"string name\" as alias")].to_vec(),
            is_default: false,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            named: [String::from("export1"), String::from("/* … */")].to_vec(),
            is_default: true,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            named: Vec::new(),
            is_default: true,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            named: Vec::new(),
            is_default: false,
            line: 0,
        },
    ];
    for (i, import) in imports.iter().enumerate() {
        assert_eq!(import.source, expected[i].source);
        assert_eq!(import.named, expected[i].named);
        assert_eq!(import.is_default, expected[i].is_default);
    }
    Ok(())
}

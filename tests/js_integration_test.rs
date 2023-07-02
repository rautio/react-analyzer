use react_analyzer::languages::javascript::JavaScript;
use react_analyzer::languages::Import;
use std::path::PathBuf;

#[test]
fn test_parse_module() -> Result<(), Box<dyn std::error::Error>> {
    let lang = JavaScript {};
    let mut path = PathBuf::from(env!("CARGO_MANIFEST_DIR"));
    path.push("tests/mocks/import.js");
    let p = path.as_path();
    let (imports, _) = lang.parse_module(p);
    let expected = [
        Import {
            source: String::from("\"module-name\""),
            default: String::from(""),
            named: Vec::new(),
            is_default: true,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            default: String::from(""),
            named: Vec::new(),
            is_default: true,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            default: String::from(""),
            named: [String::from("export1")].to_vec(),
            is_default: false,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            default: String::from(""),
            named: [String::from("export1 as alias1")].to_vec(), // Wrong for now
            is_default: true,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            default: String::from(""),
            named: [String::from("default as alias")].to_vec(), // Wrong for now
            is_default: true,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            default: String::from(""),
            named: [String::from("export1"), String::from("export2")].to_vec(),
            is_default: false,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            default: String::from(""),
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
            default: String::from(""),
            named: [String::from("\"string name\" as alias")].to_vec(),
            is_default: false,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            default: String::from(""),
            named: [String::from("export1"), String::from("/* … */")].to_vec(),
            is_default: false,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            default: String::from(""),
            named: Vec::new(),
            is_default: true,
            line: 0,
        },
        Import {
            source: String::from("\"module-name\""),
            default: String::from(""),
            named: Vec::new(),
            is_default: false,
            line: 0,
        },
    ];
    for (i, import) in imports.iter().enumerate() {
        println!("asserting: {:?}", import);
        println!("expected: {:?}", expected[i]);
        assert_eq!(import.source, expected[i].source);
        assert_eq!(import.named, expected[i].named);
        assert_eq!(import.default, expected[i].default);
    }
    Ok(())
}

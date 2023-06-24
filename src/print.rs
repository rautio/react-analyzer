use regex::Regex;
use std::path::Path;
struct Inputs {
    root: String,
    pattern: Regex,
    ignore_pattern: Regex,
    test_pattern: Regex,
}
impl std::fmt::Display for Inputs {
    fn fmt(&self, f: &mut std::fmt::Formatter) -> std::fmt::Result {
        write!(
            f,
            "Analyzing:   {}\nScan:        {}\nIgnore:      {}\nTest:        {}",
            self.root, self.pattern, self.ignore_pattern, self.test_pattern,
        )
    }
}
pub fn welcome_message() {
    println!("\n#  ....................._..................");
    println!("#  ._.__.___..__._..___|.|_................");
    println!("#  |.'__/._.\\/._`.|/.__|.__|...............");
    println!("#  |.|.|..__/.(_|.|.(__|.|_................");
    println!("#  |_|..\\___|\\__,_|\\___|\\__|...............");
    println!("#  ........................................");
    println!("#  ..................._....................");
    println!("#  ..__._._.__...__._|.|_..._._______._.__.");
    println!("#  ./._`.|.'_.\\./._`.|.|.|.|.|_../._.\\.'__|");
    println!("#  |.(_|.|.|.|.|.(_|.|.|.|_|.|/./..__/.|...");
    println!("#  .\\__,_|_|.|_|\\__,_|_|\\__,./___\\___|_|...");
    println!("#  .....................|___/..............\n");
}

pub fn input(root: &Path, pattern: Regex, ignore_pattern: Regex, test_pattern: Regex) {
    println!(
        "{}\n",
        Inputs {
            root: root.display().to_string(),
            pattern,
            ignore_pattern,
            test_pattern,
        }
    )
}

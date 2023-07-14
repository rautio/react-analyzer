use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;
use std::path::PathBuf;

#[derive(Serialize, Deserialize, Debug, PartialEq)]
#[serde(rename_all = "camelCase")]
pub struct PackageJson {
    pub dependencies: Option<HashMap<String, String>>,
    pub dev_dependencies: Option<HashMap<String, String>>,
    pub peer_dependencies: Option<HashMap<String, String>>,
    pub file_path: PathBuf,
}

pub fn parse(package_jsons: Vec<PathBuf>) -> Vec<PackageJson> {
    let mut result: Vec<PackageJson> = Vec::new();
    for p_json in package_jsons {
        let file_string = fs::read_to_string(&p_json).expect(&format!(
            "Unable to read file: {}",
            &p_json.display().to_string()
        ));
        let mut parsed_p_json: PackageJson =
            serde_json::from_str(file_string.as_str()).expect(&format!(
                "JSON was not well-formatted in: {}",
                &p_json.display().to_string()
            ));
        parsed_p_json.file_path = p_json;
        result.push(parsed_p_json)
    }
    return result;
}

pub fn list_dependencies(package_json: PackageJson) -> Vec<String> {
    let mut dependencies: Vec<String> = Vec::new();
    if package_json.dependencies.is_some() {
        let keys = &mut package_json
            .dependencies
            .unwrap()
            .keys()
            .cloned()
            .collect::<Vec<String>>();
        dependencies.append(keys);
    }
    return dependencies;
}

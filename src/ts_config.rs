use super::path_utils::path_distance;
use serde::{Deserialize, Serialize};
use serde_jsonrc;
use std::collections::HashMap;
use std::fs;
use std::path::PathBuf;

#[derive(Serialize, Deserialize, Debug, Clone, PartialEq)]
#[serde(rename_all = "camelCase")]
pub struct CompilerOptions {
    pub base_url: Option<String>,
    pub paths: Option<HashMap<String, Vec<String>>>,
}

#[derive(Serialize, Deserialize, Debug, Clone, PartialEq)]
#[serde(rename_all = "camelCase")]
pub struct TypeScriptConfig {
    pub compiler_options: Option<CompilerOptions>,
    pub file_path: Option<String>,
}

pub fn parse(ts_configs: Vec<PathBuf>, root_prefix: PathBuf) -> Vec<TypeScriptConfig> {
    let mut result: Vec<TypeScriptConfig> = Vec::new();
    for ts_config in ts_configs {
        let file_string = fs::read_to_string(&ts_config).expect(&format!(
            "Unable to read file: {}",
            &ts_config.display().to_string()
        ));
        let mut parsed_ts_config: TypeScriptConfig = serde_jsonrc::from_str(file_string.as_str())
            .expect(&format!(
                "JSON was not well-formatted in: {}",
                &ts_config.display().to_string()
            ));
        // Remove root path
        parsed_ts_config.file_path = Some(
            ts_config
                .strip_prefix(&root_prefix)
                .unwrap()
                .display()
                .to_string(),
        );
        result.push(parsed_ts_config)
    }
    return result;
}

/// Get map of file import aliases from the typescript config.
// pub fn get_aliases(ts_config: TypeScriptConfig) -> HashMap<String, PathBuf> {
//     let aliases = HashMap::new();
//     return aliases;
// }

/// Given a list of ts configs and a path get the config that is applied to the given path.
/// Primarily for use in monorepo structures to make sure aliases don't cross configs.
pub fn get_closest(ts_configs: &Vec<TypeScriptConfig>, path: PathBuf) -> Option<&TypeScriptConfig> {
    let mut closest: Option<&TypeScriptConfig> = None;
    let mut closest_distance = 0;
    for config in ts_configs {
        let cur_config = config;
        let mut config_path = PathBuf::from(&cur_config.file_path.as_ref().unwrap());
        config_path.pop(); // Last component is the actual
        if path.starts_with(&config_path) {
            // let closest_distance = path_distance(closest_path.to_path_buf(), path.clone());
            // It's a match
            if closest.is_none() {
                // No closest set yet
                closest = Some(cur_config);
                closest_distance = path_distance(path.clone(), config_path.clone());
            } else if closest.is_some() {
                let config_distance = path_distance(config_path, path.clone());
                if config_distance < closest_distance {
                    // Current config is closer
                    closest = Some(cur_config);
                    closest_distance = config_distance;
                }
            }
        }
    }
    return closest;
}

pub fn get_aliases(ts_config: Option<TypeScriptConfig>) -> Option<HashMap<String, Vec<String>>> {
    return ts_config?.compiler_options?.paths;
}
pub fn get_base_path(ts_config: Option<TypeScriptConfig>) -> Option<String> {
    return ts_config?.compiler_options?.base_url;
}

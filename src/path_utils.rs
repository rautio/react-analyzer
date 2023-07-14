use std::path::PathBuf;

/// Calculate the number of directories to get to from path A to path B.
pub fn path_distance(path_a: PathBuf, path_b: PathBuf) -> usize {
    let mut a = path_a; // A is shorter
    let mut b = path_b;
    println!("a: {}", a.display());
    println!("b: {}", b.display());
    if a.cmp(&b).is_eq() {
        return 0;
    } else {
        println!("Not equal: {} vs {}", a.display(), b.display());
    }
    if a.components().count() > b.components().count() {
        a.pop();
        return 1 + path_distance(a, b);
    }
    b.pop();
    return 1 + path_distance(a, b);
}

#[cfg(test)]
mod tests {
    use super::*;
    #[test]
    fn test_path_distance() -> Result<(), String> {
        assert_eq!(path_distance(PathBuf::from("/"), PathBuf::from("/")), 0);
        assert_eq!(
            path_distance(PathBuf::from("/foo"), PathBuf::from("/foo")),
            0
        );
        assert_eq!(
            path_distance(PathBuf::from("/foo"), PathBuf::from("/foo/test.js")),
            1
        );
        assert_eq!(
            path_distance(PathBuf::from("/foo/rick.js"), PathBuf::from("/foo/roll.js")),
            2 // Siblings are further apart than parent/child
        );
        assert_eq!(
            path_distance(PathBuf::from("/foo/bar"), PathBuf::from("/foo")),
            1
        );
        assert_eq!(
            path_distance(PathBuf::from("/foo"), PathBuf::from("/foo/bar")),
            1
        );
        assert_eq!(
            path_distance(PathBuf::from("/foo"), PathBuf::from("/bar")),
            2
        );
        assert_eq!(
            path_distance(
                PathBuf::from("/foo"),
                PathBuf::from("/bar/zoo/test/rick/roll.txt")
            ),
            6
        );
        Ok(())
    }
}

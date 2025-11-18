use anyhow::{Context, Result};
use std::fs;
use std::path::PathBuf;

pub fn get_config_dir() -> Result<PathBuf> {
    let home = dirs::home_dir().context("Could not find home directory")?;
    let config_dir = home.join(".config").join("gitconfigs");
    if !config_dir.exists() {
        fs::create_dir_all(&config_dir).context("Failed to create config directory")?;
    }
    Ok(config_dir)
}

pub fn get_git_config_path() -> Result<PathBuf> {
    let home = dirs::home_dir().context("Could not find home directory")?;
    Ok(home.join(".gitconfig"))
}

pub fn hash_file(path: &std::path::Path) -> Result<String> {
    let content = fs::read(path).context(format!("Failed to read file {:?}", path))?;
    let digest = md5::compute(content);
    Ok(format!("{:x}", digest))
}

#[cfg(unix)]
pub fn create_symlink<P: AsRef<std::path::Path>, Q: AsRef<std::path::Path>>(original: P, link: Q) -> std::io::Result<()> {
    std::os::unix::fs::symlink(original, link)
}

#[cfg(windows)]
pub fn create_symlink<P: AsRef<std::path::Path>, Q: AsRef<std::path::Path>>(original: P, link: Q) -> std::io::Result<()> {
    std::os::windows::fs::symlink_file(original, link)
}

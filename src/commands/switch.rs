use crate::utils;
use anyhow::{Context, Result};
use colored::*;
use std::fs;

pub fn execute(profile_name: &str) {
    if let Err(e) = run(profile_name) {
        eprintln!("{} {}", "Error:".red(), e);
        std::process::exit(1);
    }
}

fn run(profile_name: &str) -> Result<()> {
    let config_dir = utils::get_config_dir()?;
    let git_config_path = utils::get_git_config_path()?;

    let target_profile_path = config_dir.join(profile_name);

    if !target_profile_path.exists() {
        eprintln!(
            "{} Profile {:?} does not exist at {:?}",
            "Error:".red(),
            profile_name,
            target_profile_path
        );
        // List available profiles
        println!("Available profiles:");
        if let Ok(entries) = fs::read_dir(&config_dir) {
            for entry in entries.flatten() {
                if let Some(name) = entry.file_name().to_str() {
                    if !name.starts_with('.') {
                        println!("  - {}", name);
                    }
                }
            }
        }
        std::process::exit(1);
    }

    // Remove current .gitconfig if it exists
    if git_config_path.exists() || fs::symlink_metadata(&git_config_path).is_ok() {
        fs::remove_file(&git_config_path).context("Failed to remove existing .gitconfig")?;
    }

    // Create symlink
    utils::create_symlink(&target_profile_path, &git_config_path).context("Failed to create symlink")?;

    println!(
        "{} Switched to profile {:?}. ~/.gitconfig now points to {:?}.",
        "Success:".blue(),
        profile_name,
        target_profile_path
    );

    Ok(())
}

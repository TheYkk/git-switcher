use crate::utils;
use anyhow::{Context, Result};
use colored::*;
use std::fs;

pub fn execute() {
    if let Err(e) = run() {
        eprintln!("{} {}", "Error:".red(), e);
        std::process::exit(1);
    }
}

fn run() -> Result<()> {
    let config_dir = utils::get_config_dir()?;
    let git_config_path = utils::get_git_config_path()?;

    let mut profiles = Vec::new();
    for entry in fs::read_dir(&config_dir).context("Failed to read config directory")? {
        let entry = entry?;
        let path = entry.path();
        if path.is_file() {
            if let Some(name) = path.file_name().and_then(|n| n.to_str()) {
                if !name.starts_with('.') {
                    profiles.push(name.to_string());
                }
            }
        }
    }

    if profiles.is_empty() {
        println!("No git configuration profiles found in {:?}.", config_dir);
        println!("You can create your first profile using 'git-switcher create'.");
        return Ok(());
    }

    profiles.sort();

    let mut current_profile = None;
    if git_config_path.exists() {
        // Check if it's a symlink first
        if let Ok(target) = fs::read_link(&git_config_path) {
             if let Some(name) = target.file_name().and_then(|n| n.to_str()) {
                 current_profile = Some(name.to_string());
             }
        } else {
            // Fallback to hash check if not a symlink or read_link failed (e.g. regular file)
             if let Ok(hash) = utils::hash_file(&git_config_path) {
                for profile in &profiles {
                    let profile_path = config_dir.join(profile);
                    if let Ok(p_hash) = utils::hash_file(&profile_path) {
                        if hash == p_hash {
                            current_profile = Some(profile.clone());
                            break;
                        }
                    }
                }
            }
        }
    }

    println!("Available git configuration profiles in {:?}:\n", config_dir);

    for profile in &profiles {
        if Some(profile) == current_profile.as_ref() {
            println!("{} {} (current)", "*".green(), profile.green());
        } else {
            println!("  {}", profile);
        }
    }

    if current_profile.is_none() {
        println!("\n{}", "No active profile detected or current .gitconfig is not managed by git-switcher.".yellow());
        println!("Use 'git-switcher switch <profile>' to activate a profile.");
    }

    Ok(())
}

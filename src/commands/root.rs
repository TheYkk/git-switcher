use crate::commands::switch;
use crate::utils;
use anyhow::{Context, Result};
use colored::*;
use inquire::Select;
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

    // 1. Ensure config dir exists (handled by get_config_dir)
    // 2. Check if .gitconfig exists and back it up if not managed
    if git_config_path.exists() {
        let is_symlink = fs::symlink_metadata(&git_config_path)?
            .file_type()
            .is_symlink();
        if !is_symlink {
            // Check if it's already backed up or matches a profile
            let hash = utils::hash_file(&git_config_path)?;
            let mut matches_profile = false;

            // Check if any profile matches this hash
            if let Ok(entries) = fs::read_dir(&config_dir) {
                for entry in entries.flatten() {
                    if let Ok(p_hash) = utils::hash_file(&entry.path()) {
                        if hash == p_hash {
                            matches_profile = true;
                            break;
                        }
                    }
                }
            }

            if !matches_profile {
                let backup_path = config_dir.join("old-configs");
                if !backup_path.exists() {
                    fs::hard_link(&git_config_path, &backup_path)
                        .or_else(|_| fs::copy(&git_config_path, &backup_path).map(|_| ()))
                        .context("Failed to backup current .gitconfig")?;
                    println!(
                        "{} Current .gitconfig backed up to {:?}",
                        "Info:".blue(),
                        backup_path
                    );
                } else {
                    println!(
                        "{} {:?} already exists. Current .gitconfig not linked as old-configs.",
                        "Info:".blue(),
                        backup_path
                    );
                }
            }
        }
    } else {
        // Create default if not exists
        fs::write(&git_config_path, "[user]\n\tname = username")
            .context("Failed to create default .gitconfig")?;
    }

    // 3. List profiles
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
        println!("You can create one using 'git-switcher create'.");
        return Ok(());
    }

    profiles.sort();

    // 4. Identify current profile
    let mut current_index = 0;
    let mut current_profile_name = "unknown".to_string();

    if git_config_path.exists() {
        if let Ok(target) = fs::read_link(&git_config_path) {
            if let Some(name) = target.file_name().and_then(|n| n.to_str()) {
                if let Some(idx) = profiles.iter().position(|p| p == name) {
                    current_index = idx;
                    current_profile_name = name.to_string();
                }
            }
        } else if let Ok(hash) = utils::hash_file(&git_config_path) {
            for (i, profile) in profiles.iter().enumerate() {
                let profile_path = config_dir.join(profile);
                if let Ok(p_hash) = utils::hash_file(&profile_path) {
                    if hash == p_hash {
                        current_index = i;
                        current_profile_name = profile.clone();
                        break;
                    }
                }
            }
        }
    }

    // 5. Prompt
    let selection = Select::new(
        &format!("Select Git Config (Current: {})", current_profile_name),
        profiles,
    )
    .with_starting_cursor(current_index)
    .prompt();

    match selection {
        Ok(choice) => {
            switch::execute(&choice);
        }
        Err(_) => {
            println!("Operation cancelled.");
        }
    }

    Ok(())
}

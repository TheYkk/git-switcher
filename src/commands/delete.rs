use crate::utils;
use anyhow::{Context, Result};
use colored::*;
use inquire::{Confirm, Select};
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

    // List profiles
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
        println!("No git configuration profiles found to delete in {:?}.", config_dir);
        return Ok(());
    }

    profiles.sort();

    // Identify current profile
    let mut current_index = 0;
    let mut current_profile_name = "none".to_string();

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

    // Select profile to delete
    let selection = Select::new(
        &format!("Select Git Config profile to delete (Current: {})", current_profile_name),
        profiles,
    )
    .with_starting_cursor(current_index)
    .prompt()
    .context("Prompt failed")?;

    // Confirm
    let confirmed = Confirm::new(&format!("Are you sure you want to delete profile {:?}?", selection))
        .with_default(false)
        .prompt()
        .context("Confirmation failed")?;

    if !confirmed {
        println!("{} Profile {:?} not deleted.", "Info:".blue(), selection);
        return Ok(());
    }

    // Delete
    let profile_path = config_dir.join(&selection);
    fs::remove_file(&profile_path).context("Failed to delete profile file")?;

    // If active, remove symlink
    if selection == current_profile_name {
        if git_config_path.exists() || fs::symlink_metadata(&git_config_path).is_ok() {
            fs::remove_file(&git_config_path).context("Failed to remove current .gitconfig symlink")?;
        }
        println!("{} Profile {:?} deleted. Current .gitconfig was also removed.", "Success:".green(), selection);
    } else {
        println!("{} Profile {:?} deleted.", "Success:".green(), selection);
    }

    Ok(())
}

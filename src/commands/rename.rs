use crate::utils;
use anyhow::{Context, Result};
use colored::*;
use inquire::{Select, Text};
use std::fs;
use std::os::unix::fs::symlink;

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
        println!(
            "No git configuration profiles found to rename in {:?}.",
            config_dir
        );
        return Ok(());
    }

    profiles.sort();

    // Select profile
    let old_name = Select::new("Select profile to rename", profiles.clone())
        .prompt()
        .context("Prompt failed")?;

    // New name
    let old_name_clone = old_name.clone();
    let new_name = Text::new(&format!("Enter new name for profile {:?}", old_name))
        .with_validator(move |input: &str| {
            if input.is_empty() {
                Ok(inquire::validator::Validation::Invalid(
                    "Name cannot be empty".into(),
                ))
            } else if profiles.contains(&input.to_string()) && input != old_name_clone {
                Ok(inquire::validator::Validation::Invalid(
                    format!("Profile {:?} already exists", input).into(),
                ))
            } else {
                Ok(inquire::validator::Validation::Valid)
            }
        })
        .prompt()
        .context("Prompt failed")?;

    if old_name == new_name {
        println!(
            "{}",
            "New name is the same as the old name. No changes made.".yellow()
        );
        return Ok(());
    }

    let old_path = config_dir.join(&old_name);
    let new_path = config_dir.join(&new_name);

    fs::rename(&old_path, &new_path).context("Failed to rename profile")?;
    println!(
        "{} Profile {:?} renamed to {:?}.",
        "Success:".green(),
        old_name,
        new_name
    );

    // Update symlink if active
    if git_config_path.exists() || fs::symlink_metadata(&git_config_path).is_ok() {
        if let Ok(target) = fs::read_link(&git_config_path) {
            if target == old_path {
                fs::remove_file(&git_config_path).context("Failed to remove old symlink")?;
                symlink(&new_path, &git_config_path).context("Failed to create new symlink")?;
                println!(
                    "{} Active profile symlink updated to {:?}.",
                    "Info:".blue(),
                    new_name
                );
            }
        }
    }

    Ok(())
}

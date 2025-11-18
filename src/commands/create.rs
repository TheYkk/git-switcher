use crate::utils;
use anyhow::{Context, Result};
use colored::*;
use inquire::Text;
use std::fs;

pub fn execute() {
    if let Err(e) = run() {
        eprintln!("{} {}", "Error:".red(), e);
        std::process::exit(1);
    }
}

fn run() -> Result<()> {
    let config_dir = utils::get_config_dir()?;

    let name = Text::new("Profile name:")
        .with_validator(|input: &str| {
            if input.is_empty() {
                Ok(inquire::validator::Validation::Invalid(
                    "Name cannot be empty".into(),
                ))
            } else {
                Ok(inquire::validator::Validation::Valid)
            }
        })
        .prompt()
        .context("Failed to get profile name")?;

    let profile_path = config_dir.join(&name);

    if profile_path.exists() {
        println!(
            "{} Profile {:?} already exists at {:?}",
            "Error:".red(),
            name,
            profile_path
        );
        return Ok(());
    }

    let content = format!(
        "[user]\n\tname = {}\n\temail = your_email@example.com",
        name
    );
    fs::write(&profile_path, content).context("Failed to write profile file")?;

    println!(
        "{} Profile {:?} created successfully at {:?}",
        "Success:".green(),
        name,
        profile_path
    );
    println!(
        "{}",
        "Please edit the file to set your desired git user name and email.".yellow()
    );

    Ok(())
}

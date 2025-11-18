use crate::utils;
use anyhow::{Context, Result};
use colored::*;
use std::env;
use std::process::Command;

pub fn execute() {
    if let Err(e) = run() {
        eprintln!("{} {}", "Error:".red(), e);
        std::process::exit(1);
    }
}

fn run() -> Result<()> {
    let git_config_path = utils::get_git_config_path()?;

    if !git_config_path.exists() {
        println!("{} No active .gitconfig found at {:?} to edit.", "Warning:".yellow(), git_config_path);
        println!("{}", "Consider switching to or creating a profile first.".yellow());
        return Ok(());
    }

    let editor = env::var("EDITOR").unwrap_or_else(|_| "vim".to_string());
    
    // Simple split by space, might not handle quotes perfectly but sufficient for most EDITOR vars
    // For more robust handling, shell-words crate could be used, but keeping deps minimal
    let parts: Vec<&str> = editor.split_whitespace().collect();
    if parts.is_empty() {
        anyhow::bail!("EDITOR environment variable is empty");
    }

    let mut cmd = Command::new(parts[0]);
    if parts.len() > 1 {
        cmd.args(&parts[1..]);
    }
    cmd.arg(&git_config_path);

    println!("{} Opening {:?} with {}...", "Info:".blue(), git_config_path, editor);

    let status = cmd.status().context(format!("Failed to run editor {}", editor))?;

    if !status.success() {
        eprintln!("{} Editor exited with non-zero status", "Error:".red());
    }

    Ok(())
}

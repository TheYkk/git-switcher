mod commands;
mod utils;

use clap::{Parser, Subcommand};
use commands::*;

#[derive(Parser)]
#[command(name = "git-switcher")]
#[command(about = "A tool to easily switch between different git configurations", long_about = None)]
struct Cli {
    #[command(subcommand)]
    command: Option<Commands>,
}

#[derive(Subcommand)]
enum Commands {
    /// Creates a new git configuration profile
    #[command(
        long_about = "Creates a new git configuration profile.\nYou will be prompted to enter a name for the new profile.\nA new configuration file will be created in the ~/.config/gitconfigs directory."
    )]
    Create,
    /// Deletes an existing git configuration profile
    #[command(
        long_about = "Deletes an existing git configuration profile.\nYou will be prompted to select a profile to delete from the available profiles.\nThe selected configuration file will be removed from the ~/.config/gitconfigs directory.\nIf the deleted profile is the currently active one, ~/.gitconfig will also be removed."
    )]
    Delete,
    /// Opens the current ~/.gitconfig file in your default editor
    #[command(
        long_about = "Opens the currently active ~/.gitconfig file in your system's default editor (or $EDITOR environment variable if set).\nThis allows you to directly modify the active git configuration."
    )]
    Edit,
    /// Lists all available git configuration profiles
    #[command(
        long_about = "Lists all available git configuration profiles stored in ~/.config/gitconfigs.\nThe currently active profile (if any) will be marked with an asterisk (*) and highlighted."
    )]
    List,
    /// Renames an existing git configuration profile
    #[command(
        long_about = "Renames an existing git configuration profile.\nYou will be prompted to select the profile to rename and then to enter the new name.\nThe configuration file in ~/.config/gitconfigs will be renamed.\nIf the renamed profile is the currently active one, the ~/.gitconfig symlink will be updated."
    )]
    Rename,
    /// Switches the active git configuration to the specified profile
    #[command(
        long_about = "Switches the active git configuration to the specified profile.\nThe command takes exactly one argument: the name of the profile to switch to.\nThis profile must exist in the ~/.config/gitconfigs directory.\nThe ~/.gitconfig file will be updated to be a symlink to the selected profile."
    )]
    Switch {
        /// The name of the profile to switch to
        profile_name: String,
    },
}

fn main() {
    let cli = Cli::parse();

    match &cli.command {
        Some(Commands::Create) => create::execute(),
        Some(Commands::Delete) => delete::execute(),
        Some(Commands::Edit) => edit::execute(),
        Some(Commands::List) => list::execute(),
        Some(Commands::Rename) => rename::execute(),
        Some(Commands::Switch { profile_name }) => switch::execute(profile_name),
        None => root::execute(),
    }
}

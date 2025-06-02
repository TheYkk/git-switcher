<p align="center">
  <img height="300" src="https://user-images.githubusercontent.com/53150440/135754337-e9a12311-9eb0-4de6-8d19-348341f3b200.png"/><br/>
  <a>
        <img src="https://github.com/TheYkk/git-switcher/actions/workflows/release.yml/badge.svg">
  </a>
  <a>
        <img src="https://img.shields.io/github/v/release/theykk/git-switcher?style=flat&labelColor=1C2C2E&color=abc3d6&logo=GitHub&logoColor=white">
  </a>
  <a>
        <img src="https://img.shields.io/github/license/theykk/git-switcher?style=flat&labelColor=1C2C2E&color=abc3d6&logoColor=white">
  </a>
  <a>
        <img src="https://img.shields.io/github/stars/theykk/git-switcher?style=flat&labelColor=1C2C2E&color=abc3d6&logoColor=white">
  </a>
  
</p>

# Git Profile Switcher

Switch between your git profiles easily

## Install

With Brew

```
brew install theykk/tap/git-switcher
```

With golang

```
go install github.com/theykk/git-switcher@latest
```

With AUR
```
yay -S git-switcher
```
or you can install like this:
```
git clone https://aur.archlinux.org/git-switcher.git
makepkg -is 
```

## Commands

Git Switcher offers several commands to help you manage your Git profiles:

### `list`

Lists all your saved git profiles. The currently active profile will be highlighted with an asterisk (*) and marked as `(current)`.

**Usage:**

```sh
git-switcher list
```

**Example Output:**

```
Available Git profiles:
  work-profile
* personal-profile (current)
  freelance-project
```

### `switch`

Allows you to interactively select and switch to a different Git profile from your saved list. This is also the default behavior when running `git-switcher` without any subcommand.

*(The GIF below demonstrates this functionality)*

### `create`

Guides you through the process of creating and saving a new Git profile.

*(The GIF below demonstrates this functionality)*

### `delete`

Allows you to select and delete one of your saved Git profiles.

*(The GIF below demonstrates this functionality)*

### `rename`

Allows you to rename an existing saved Git profile.

*(The GIF below demonstrates this functionality)*

## Switch Profile

![Switcher](https://user-images.githubusercontent.com/53150440/135753964-94d83bf5-597c-4983-b0cf-5da6f12e6c7c.gif)

## Create Profile

![Create](https://user-images.githubusercontent.com/53150440/135753994-aa60050b-020c-422b-9fed-0a266f550dda.gif)

## Delete Profile

![Delete](https://user-images.githubusercontent.com/53150440/135754022-55268cc5-9979-4802-8a93-0e09c158cd6c.gif)

## Rename Profile

![Rename](https://user-images.githubusercontent.com/53150440/135754365-f8e347d9-89e1-4a34-a131-edeb7e004047.gif)

## Feedback

If you have any feedback, please reach out to us at yusufkaan142@gmail.com

## License

[Apache-2.0](https://choosealicense.com/licenses/apache-2.0/)

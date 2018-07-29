# Contributing to THORChain

:+1::tada: First off, thanks for taking the time to contribute! :tada::+1:

The following is a set of guidelines for contributing to THORChain and its packages, which are hosted in the [THORChain Organization](https://github.com/thorchain) on GitHub. These are mostly guidelines, not rules. Use your best judgment, and feel free to propose changes to this document in a pull request.

#### Table Of Contents

[Code of Conduct](#code-of-conduct)

[I don't want to read this whole thing, I just have a question!!!](#i-dont-want-to-read-this-whole-thing-i-just-have-a-question)

[What should I know before I get started?](#what-should-i-know-before-i-get-started)
  * [THORChain and Packages](#thorchain-and-packages)
  * [THORChain Design Decisions](#design-decisions)

[How Can I Contribute?](#how-can-i-contribute)
  * [Reporting Bugs](#reporting-bugs)
  * [Suggesting Enhancements](#suggesting-enhancements)
  * [Your First Code Contribution](#your-first-code-contribution)
  * [Pull Requests](#pull-requests)

[Styleguides](#styleguides)
  * [Git Commit Messages](#git-commit-messages)
  * [JavaScript Styleguide](#javascript-styleguide)
  * [CoffeeScript Styleguide](#coffeescript-styleguide)
  * [Specs Styleguide](#specs-styleguide)
  * [Documentation Styleguide](#documentation-styleguide)

[Additional Notes](#additional-notes)
  * [Issue and Pull Request Labels](#issue-and-pull-request-labels)

## Code of Conduct

This project and everyone participating in it is governed by the [THORChain Code of Conduct](CODE_OF_CONDUCT.md). By participating, you are expected to uphold this code. Please report unacceptable behavior to [devs@thorchain.org](mailto:devs@thorchain.org).

## I don't want to read this whole thing I just have a question!!!

> **Note:** [Please don't file an issue to ask a question.](https://blog.thorchain.com) You'll get faster results by using the resources below.

We have an official message board with a detailed FAQ and where the community chimes in with helpful advice if you have questions.

* [The official THORChain message board](https://thorchain.org)
* [THORChain FAQ](https://thorchain.org)

If chat is more your speed, you can join the THORChain Discord

* [Join the THORChain Team](https://thorchain.org)
    * Even though Slack is a chat service, sometimes it takes several hours for community members to respond &mdash; please be patient!
    * Use the `#thorchain` channel for general questions or discussion about THORChain
    * Use the `#bifrost` channel for questions about Bifrost
    * Use the `#ui` channel for questions and discussion about THORChain UI and themes
    * There are many other channels available, check the channel list

## What should I know before I get started?

### THORChain 

THORChain is a large open source project &mdash; it's made up of over [10 repositories](https://github.com/thorchain). When you initially consider contributing to THORChain, you might be unsure about which of those 10 repositories implements the functionality you want to change or report a bug for. This section should help you with that.

Here's a list of the big ones:

* [THORChain](https://github.com/thorchain/THORChain)

There are many more, but this list should be a good starting point.

Also, because THORChain is so extensible, it's possible that a feature you've become accustomed to in THORChain or an issue you're encountering isn't coming from a bundled package at all, but rather a [community package](https://thorchain.io/packages) you've installed. Each community package has its own repository too, the [THORChain FAQ](https://discuss.thorchain.com/c/faq) has instructions on how to contact: 

#### Package Conventions

There are a few conventions that have developed over time around packages:

* Packages that add one or more syntax highlighting grammars are named `language-[language-name]`
    * Language packages can add other things besides just a grammar. Many offer commonly-used snippets. Try not to add too much though.
* Theme packages are split into two categories: UI and Syntax themes
    * UI themes are named `[theme-name]-ui`
    * Syntax themes are named `[theme-name]-syntax`
    * Often themes that are designed to work together are given the same root name, for example: `one-dark-ui` and `one-dark-syntax`
    * UI themes style everything outside of the editor pane &mdash; all of the green areas in the [packages image above](#thorchain-packages-image)
    * Syntax themes style just the items inside the editor pane, mostly syntax highlighting


### Design Decisions

When we make a significant decision in how we maintain the project and what we can or cannot support, we will document it in the [thorchain/design-decisions repository](https://github.com/thorchain/design-decisions). If you have a question around how we do things, check to see if it is documented there. If it is *not* documented there, please open a new topic on [Disqus, the official THORChain message board](https://disqus.thorchain.com) and ask your question.

## How Can I Contribute?

### Reporting Bugs

This section guides you through submitting a bug report for THORChain. Following these guidelines helps maintainers and the community understand your report :pencil:, reproduce the behavior :computer: :computer:, and find related reports :mag_right:.

Before creating bug reports, please check [this list](#before-submitting-a-bug-report) as you might find out that you don't need to create one. When you are creating a bug report, please [include as many details as possible](#how-do-i-submit-a-good-bug-report). Fill out [the required template](ISSUE_TEMPLATE.md), the information it asks for helps us resolve issues faster.

> **Note:** If you find a **Closed** issue that seems like it is the same thing that you're experiencing, open a new issue and include a link to the original issue in the body of your new one.

#### Before Submitting A Bug Report

* **Check the [debugging guide](https://flight-manual.thorchain.io/hacking-thorchain/sections/debugging/).** You might be able to find the cause of the problem and fix things yourself. Most importantly, check if you can reproduce the problem [in the latest version of THORChain](https://flight-manual.thorchain.io/hacking-thorchain/sections/debugging/#update-to-the-latest-version), if the problem happens when you run THORChain in [safe mode](https://flight-manual.thorchain.io/hacking-thorchain/sections/debugging/#check-if-the-problem-shows-up-in-safe-mode), and if you can get the desired behavior by changing [THORChain's or packages' config settings](https://flight-manual.thorchain.io/hacking-thorchain/sections/debugging/#check-thorchain-and-package-settings).
* **Check the [FAQs on the forum](https://discuss.thorchain.io/c/faq)** for a list of common questions and problems.
* **Determine [which repository the problem should be reported in](#thorchain-and-packages)**.
* **Perform a [cursory search](https://github.com/search?q=+is%3Aissue+user%3Athorchain)** to see if the problem has already been reported. If it has **and the issue is still open**, add a comment to the existing issue instead of opening a new one.

#### How Do I Submit A (Good) Bug Report?

Bugs are tracked as [GitHub issues](https://guides.github.com/features/issues/). After you've determined [which repository](#thorchain-and-packages) your bug is related to, create an issue on that repository and provide the following information by filling in [the template](ISSUE_TEMPLATE.md).

Explain the problem and include additional details to help maintainers reproduce the problem:

* **Use a clear and descriptive title** for the issue to identify the problem.
* **Describe the exact steps which reproduce the problem** in as many details as possible. For example, start by explaining how you started THORChain, e.g. which command exactly you used in the terminal, or how you started THORChain otherwise. When listing steps, **don't just say what you did, but explain how you did it**. For example, if you moved the cursor to the end of a line, explain if you used the mouse, or a keyboard shortcut or an THORChain command, and if so which one?
* **Provide specific examples to demonstrate the steps**. Include links to files or GitHub projects, or copy/pasteable snippets, which you use in those examples. If you're providing snippets in the issue, use [Markdown code blocks](https://help.github.com/articles/markdown-basics/#multiple-lines).
* **Describe the behavior you observed after following the steps** and point out what exactly is the problem with that behavior.
* **Explain which behavior you expected to see instead and why.**
* **Include screenshots and animated GIFs** which show you following the described steps and clearly demonstrate the problem. If you use the keyboard while following the steps, **record the GIF with the [Keybinding Resolver](https://github.com/thorchain/keybinding-resolver) shown**. You can use [this tool](https://www.cockos.com/licecap/) to record GIFs on macOS and Windows, and [this tool](https://github.com/colinkeenan/silentcast) or [this tool](https://github.com/GNOME/byzanz) on Linux.
* **If you're reporting that THORChain crashed**, include a crash report with a stack trace from the operating system. On macOS, the crash report will be available in `Console.app` under "Diagnostic and usage information" > "User diagnostic reports". Include the crash report in the issue in a [code block](https://help.github.com/articles/markdown-basics/#multiple-lines), a [file attachment](https://help.github.com/articles/file-attachments-on-issues-and-pull-requests/), or put it in a [gist](https://gist.github.com/) and provide link to that gist.
* **If the problem is related to performance or memory**, include a [CPU profile capture](https://flight-manual.thorchain.io/hacking-thorchain/sections/debugging/#diagnose-runtime-performance) with your report.
* **If Chrome's developer tools pane is shown without you triggering it**, that normally means that you have a syntax error in one of your themes or in your `styles.less`. Try running in [Safe Mode](https://flight-manual.thorchain.io/hacking-thorchain/sections/debugging/#using-safe-mode) and using a different theme or comment out the contents of your `styles.less` to see if that fixes the problem.
* **If the problem wasn't triggered by a specific action**, describe what you were doing before the problem happened and share more information using the guidelines below.

Provide more context by answering these questions:

* **Can you reproduce the problem in [safe mode](https://flight-manual.thorchain.io/hacking-thorchain/sections/debugging/#diagnose-runtime-performance-problems-with-the-dev-tools-cpu-profiler)?**
* **Did the problem start happening recently** (e.g. after updating to a new version of THORChain) or was this always a problem?
* If the problem started happening recently, **can you reproduce the problem in an older version of THORChain?** What's the most recent version in which the problem doesn't happen? You can download older versions of THORChain from [the releases page](https://github.com/thorchain/thorchain/releases).
* **Can you reliably reproduce the issue?** If not, provide details about how often the problem happens and under which conditions it normally happens.
* If the problem is related to working with files (e.g. opening and editing files), **does the problem happen for all files and projects or only some?** Does the problem happen only when working with local or remote files (e.g. on network drives), with files of a specific type (e.g. only JavaScript or Python files), with large files or files with very long lines, or with files in a specific encoding? Is there anything else special about the files you are using?

Include details about your configuration and environment:

* **Which version of THORChain are you using?** You can get the exact version by running `thorchain -v` in your terminal, or by starting THORChain and running the `Application: About` command from the [Command Palette](https://github.com/thorchain/command-palette).
* **What's the name and version of the OS you're using**?
* **Are you running THORChain in a virtual machine?** If so, which VM software are you using and which operating systems and versions are used for the host and the guest?
* **Which [packages](#thorchain-and-packages) do you have installed?** You can get that list by running `apm list --installed`.
* **Are you using [local configuration files](https://flight-manual.thorchain.io/using-thorchain/sections/basic-customization/)** `config.cson`, `keymap.cson`, `snippets.cson`, `styles.less` and `init.coffee` to customize THORChain? If so, provide the contents of those files, preferably in a [code block](https://help.github.com/articles/markdown-basics/#multiple-lines) or with a link to a [gist](https://gist.github.com/).
* **Are you using THORChain with multiple monitors?** If so, can you reproduce the problem when you use a single monitor?
* **Which keyboard layout are you using?** Are you using a US layout or some other layout?

### Suggesting Enhancements

This section guides you through submitting an enhancement suggestion for THORChain, including completely new features and minor improvements to existing functionality. Following these guidelines helps maintainers and the community understand your suggestion :pencil: and find related suggestions :mag_right:.

Before creating enhancement suggestions, please check [this list](#before-submitting-an-enhancement-suggestion) as you might find out that you don't need to create one. When you are creating an enhancement suggestion, please [include as many details as possible](#how-do-i-submit-a-good-enhancement-suggestion). Fill in [the template](ISSUE_TEMPLATE.md), including the steps that you imagine you would take if the feature you're requesting existed.

#### Before Submitting An Enhancement Suggestion

* **Check the [debugging guide](https://flight-manual.thorchain.io/hacking-thorchain/sections/debugging/)** for tips — you might discover that the enhancement is already available. Most importantly, check if you're using [the latest version of THORChain](https://flight-manual.thorchain.io/hacking-thorchain/sections/debugging/#update-to-the-latest-version) and if you can get the desired behavior by changing [THORChain's or packages' config settings](https://flight-manual.thorchain.io/hacking-thorchain/sections/debugging/#check-thorchain-and-package-settings).
* **Check if there's already [a package](https://thorchain.io/packages) which provides that enhancement.**
* **Determine [which repository the enhancement should be suggested in](#thorchain-and-packages).**
* **Perform a [cursory search](https://github.com/search?q=+is%3Aissue+user%3Athorchain)** to see if the enhancement has already been suggested. If it has, add a comment to the existing issue instead of opening a new one.

#### How Do I Submit A (Good) Enhancement Suggestion?

Enhancement suggestions are tracked as [GitHub issues](https://guides.github.com/features/issues/). After you've determined [which repository](#thorchain-and-packages) your enhancement suggestion is related to, create an issue on that repository and provide the following information:

* **Use a clear and descriptive title** for the issue to identify the suggestion.
* **Provide a step-by-step description of the suggested enhancement** in as many details as possible.
* **Provide specific examples to demonstrate the steps**. Include copy/pasteable snippets which you use in those examples, as [Markdown code blocks](https://help.github.com/articles/markdown-basics/#multiple-lines).
* **Describe the current behavior** and **explain which behavior you expected to see instead** and why.
* **Include screenshots and animated GIFs** which help you demonstrate the steps or point out the part of THORChain which the suggestion is related to. You can use [this tool](https://www.cockos.com/licecap/) to record GIFs on macOS and Windows, and [this tool](https://github.com/colinkeenan/silentcast) or [this tool](https://github.com/GNOME/byzanz) on Linux.
* **Explain why this enhancement would be useful** to most THORChain users and isn't something that can or should be implemented as a [community package](#thorchain-and-packages).
* **List some other text editors or applications where this enhancement exists.**
* **Specify which version of THORChain you're using.** You can get the exact version by running `thorchain -v` in your terminal, or by starting THORChain and running the `Application: About` command from the [Command Palette](https://github.com/thorchain/command-palette).
* **Specify the name and version of the OS you're using.**

### Your First Code Contribution

Unsure where to begin contributing to THORChain? You can start by looking through these `beginner` and `help-wanted` issues:

* [Beginner issues][beginner] - issues which should only require a few lines of code, and a test or two.
* [Help wanted issues][help-wanted] - issues which should be a bit more involved than `beginner` issues.

Both issue lists are sorted by total number of comments. While not perfect, number of comments is a reasonable proxy for impact a given change will have.

If you want to read about using THORChain or developing packages in THORChain, the [THORChain Flight Manual](https://flight-manual.thorchain.io) is free and available online. You can find the source to the manual in [thorchain/flight-manual.thorchain.io](https://github.com/thorchain/flight-manual.thorchain.io).

#### Local development

THORChain Core and all packages can be developed locally. For instructions on how to do this, see the following sections in the [THORChain Flight Manual](https://flight-manual.thorchain.io):

* [Hacking on THORChain Core][hacking-on-thorchain-core]
* [Contributing to Official THORChain Packages][contributing-to-official-thorchain-packages]

### Pull Requests

* Fill in [the required template](PULL_REQUEST_TEMPLATE.md)
* Do not include issue numbers in the PR title
* Include screenshots and animated GIFs in your pull request whenever possible.
* Follow the [JavaScript](#javascript-styleguide) and [CoffeeScript](#coffeescript-styleguide) styleguides.
* Include thoughtfully-worded, well-structured [Jasmine](https://jasmine.github.io/) specs in the `./spec` folder. Run them using `thorchain --test spec`. See the [Specs Styleguide](#specs-styleguide) below.
* Document new code based on the [Documentation Styleguide](#documentation-styleguide)
* End all files with a newline
* [Avoid platform-dependent code](https://flight-manual.thorchain.io/hacking-thorchain/sections/cross-platform-compatibility/)
* Place requires in the following order:
    * Built in Node Modules (such as `path`)
    * Built in THORChain and Electron Modules (such as `thorchain`, `remote`)
    * Local Modules (using relative paths)
* Place class properties in the following order:
    * Class methods and properties (methods starting with a `@` in CoffeeScript or `static` in JavaScript)
    * Instance methods and properties

## Styleguides

### Git Commit Messages

* Use the present tense ("Add feature" not "Added feature")
* Use the imperative mood ("Move cursor to..." not "Moves cursor to...")
* Limit the first line to 72 characters or less
* Reference issues and pull requests liberally after the first line
* When only changing documentation, include `[ci skip]` in the commit title
* Consider starting the commit message with an applicable emoji:
    * :art: `:art:` when improving the format/structure of the code
    * :racehorse: `:racehorse:` when improving performance
    * :non-potable_water: `:non-potable_water:` when plugging memory leaks
    * :memo: `:memo:` when writing docs
    * :penguin: `:penguin:` when fixing something on Linux
    * :apple: `:apple:` when fixing something on macOS
    * :checkered_flag: `:checkered_flag:` when fixing something on Windows
    * :bug: `:bug:` when fixing a bug
    * :fire: `:fire:` when removing code or files
    * :green_heart: `:green_heart:` when fixing the CI build
    * :white_check_mark: `:white_check_mark:` when adding tests
    * :lock: `:lock:` when dealing with security
    * :arrow_up: `:arrow_up:` when upgrading dependencies
    * :arrow_down: `:arrow_down:` when downgrading dependencies
    * :shirt: `:shirt:` when removing linter warnings

### JavaScript Styleguide

All JavaScript must adhere to [JavaScript Standard Style](https://standardjs.com/).

* Prefer the object spread operator (`{...anotherObj}`) to `Object.assign()`
* Inline `export`s with expressions whenever possible
  ```js
  // Use this:
  export default class ClassName {

  }

  // Instead of:
  class ClassName {

  }
  export default ClassName
  ```

### CoffeeScript Styleguide

* Set parameter defaults without spaces around the equal sign
    * `clear = (count=1) ->` instead of `clear = (count = 1) ->`
* Use spaces around operators
    * `count + 1` instead of `count+1`
* Use spaces after commas (unless separated by newlines)
* Use parentheses if it improves code clarity.
* Prefer alphabetic keywords to symbolic keywords:
    * `a is b` instead of `a == b`
* Avoid spaces inside the curly-braces of hash literals:
    * `{a: 1, b: 2}` instead of `{ a: 1, b: 2 }`
* Include a single line of whitespace between methods.
* Capitalize initialisms and acronyms in names, except for the first word, which
  should be lower-case:
  * `getURI` instead of `getUri`
  * `uriToOpen` instead of `URIToOpen`
* Use `slice()` to copy an array
* Add an explicit `return` when your function ends with a `for`/`while` loop and
  you don't want it to return a collected array.
* Use `this` instead of a standalone `@`
  * `return this` instead of `return @`

### Specs Styleguide

- Include thoughtfully-worded, well-structured [Jasmine](https://jasmine.github.io/) specs in the `./spec` folder.
- Treat `describe` as a noun or situation.
- Treat `it` as a statement about state or how an operation changes state.

#### Example

```coffee
describe 'a dog', ->
 it 'barks', ->
 # spec here
 describe 'when the dog is happy', ->
  it 'wags its tail', ->
  # spec here
```

### Documentation Styleguide

* Use [THORChainDoc](https://github.com/thorchain/thorchaindoc).
* Use [Markdown](https://daringfireball.net/projects/markdown).
* Reference methods and classes in markdown with the custom `{}` notation:
    * Reference classes with `{ClassName}`
    * Reference instance methods with `{ClassName::methodName}`
    * Reference class methods with `{ClassName.methodName}`

#### Example

```coffee
# Public: Disable the package with the given name.
#
# * `name`    The {String} name of the package to disable.
# * `options` (optional) The {Object} with disable options (default: {}):
#   * `trackTime`     A {Boolean}, `true` to track the amount of time taken.
#   * `ignoreErrors`  A {Boolean}, `true` to catch and ignore errors thrown.
# * `callback` The {Function} to call after the package has been disabled.
#
# Returns `undefined`.
disablePackage: (name, options, callback) ->
```

## Additional Notes

### Issue and Pull Request Labels

This section lists the labels we use to help us track and manage issues and pull requests. Most labels are used across all THORChain repositories, but some are specific to `thorchain/thorchain`.

[GitHub search](https://help.github.com/articles/searching-issues/) makes it easy to use labels for finding groups of issues or pull requests you're interested in. For example, you might be interested in [open issues across `thorchain/thorchain` and all THORChain-owned packages which are labeled as bugs, but still need to be reliably reproduced](https://github.com/search?utf8=%E2%9C%93&q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Abug+label%3Aneeds-reproduction) or perhaps [open pull requests in `thorchain/thorchain` which haven't been reviewed yet](https://github.com/search?utf8=%E2%9C%93&q=is%3Aopen+is%3Apr+repo%3Athorchain%2Fthorchain+comments%3A0). To help you find issues and pull requests, each label is listed with search links for finding open items with that label in `thorchain/thorchain` only and also across all THORChain repositories. We  encourage you to read about [other search filters](https://help.github.com/articles/searching-issues/) which will help you write more focused queries.

The labels are loosely grouped by their purpose, but it's not required that every issue have a label from every group or that an issue can't have more than one label from the same group.

Please open an issue on `thorchain/thorchain` if you have suggestions for new labels, and if you notice some labels are missing on some repositories, then please open an issue on that repository.

#### Type of Issue and Issue State

| Label name | `thorchain/thorchain` :mag_right: | `thorchain`‑org :mag_right: | Description |
| --- | --- | --- | --- |
| `enhancement` | [search][search-thorchain-repo-label-enhancement] | [search][search-thorchain-org-label-enhancement] | Feature requests. |
| `bug` | [search][search-thorchain-repo-label-bug] | [search][search-thorchain-org-label-bug] | Confirmed bugs or reports that are very likely to be bugs. |
| `question` | [search][search-thorchain-repo-label-question] | [search][search-thorchain-org-label-question] | Questions more than bug reports or feature requests (e.g. how do I do X). |
| `feedback` | [search][search-thorchain-repo-label-feedback] | [search][search-thorchain-org-label-feedback] | General feedback more than bug reports or feature requests. |
| `help-wanted` | [search][search-thorchain-repo-label-help-wanted] | [search][search-thorchain-org-label-help-wanted] | The THORChain core team would appreciate help from the community in resolving these issues. |
| `beginner` | [search][search-thorchain-repo-label-beginner] | [search][search-thorchain-org-label-beginner] | Less complex issues which would be good first issues to work on for users who want to contribute to THORChain. |
| `more-information-needed` | [search][search-thorchain-repo-label-more-information-needed] | [search][search-thorchain-org-label-more-information-needed] | More information needs to be collected about these problems or feature requests (e.g. steps to reproduce). |
| `needs-reproduction` | [search][search-thorchain-repo-label-needs-reproduction] | [search][search-thorchain-org-label-needs-reproduction] | Likely bugs, but haven't been reliably reproduced. |
| `blocked` | [search][search-thorchain-repo-label-blocked] | [search][search-thorchain-org-label-blocked] | Issues blocked on other issues. |
| `duplicate` | [search][search-thorchain-repo-label-duplicate] | [search][search-thorchain-org-label-duplicate] | Issues which are duplicates of other issues, i.e. they have been reported before. |
| `wontfix` | [search][search-thorchain-repo-label-wontfix] | [search][search-thorchain-org-label-wontfix] | The THORChain core team has decided not to fix these issues for now, either because they're working as intended or for some other reason. |
| `invalid` | [search][search-thorchain-repo-label-invalid] | [search][search-thorchain-org-label-invalid] | Issues which aren't valid (e.g. user errors). |
| `package-idea` | [search][search-thorchain-repo-label-package-idea] | [search][search-thorchain-org-label-package-idea] | Feature request which might be good candidates for new packages, instead of extending THORChain or core THORChain packages. |
| `wrong-repo` | [search][search-thorchain-repo-label-wrong-repo] | [search][search-thorchain-org-label-wrong-repo] | Issues reported on the wrong repository (e.g. a bug related to the [Settings View package](https://github.com/thorchain/settings-view) was reported on [THORChain core](https://github.com/thorchain/thorchain)). |

#### Topic Categories

| Label name | `thorchain/thorchain` :mag_right: | `thorchain`‑org :mag_right: | Description |
| --- | --- | --- | --- |
| `windows` | [search][search-thorchain-repo-label-windows] | [search][search-thorchain-org-label-windows] | Related to THORChain running on Windows. |
| `linux` | [search][search-thorchain-repo-label-linux] | [search][search-thorchain-org-label-linux] | Related to THORChain running on Linux. |
| `mac` | [search][search-thorchain-repo-label-mac] | [search][search-thorchain-org-label-mac] | Related to THORChain running on macOS. |
| `documentation` | [search][search-thorchain-repo-label-documentation] | [search][search-thorchain-org-label-documentation] | Related to any type of documentation (e.g. [API documentation](https://thorchain.io/docs/api/latest/) and the [flight manual](https://flight-manual.thorchain.io/)). |
| `performance` | [search][search-thorchain-repo-label-performance] | [search][search-thorchain-org-label-performance] | Related to performance. |
| `security` | [search][search-thorchain-repo-label-security] | [search][search-thorchain-org-label-security] | Related to security. |
| `ui` | [search][search-thorchain-repo-label-ui] | [search][search-thorchain-org-label-ui] | Related to visual design. |
| `api` | [search][search-thorchain-repo-label-api] | [search][search-thorchain-org-label-api] | Related to THORChain's public APIs. |
| `uncaught-exception` | [search][search-thorchain-repo-label-uncaught-exception] | [search][search-thorchain-org-label-uncaught-exception] | Issues about uncaught exceptions, normally created from the [Notifications package](https://github.com/thorchain/notifications). |
| `crash` | [search][search-thorchain-repo-label-crash] | [search][search-thorchain-org-label-crash] | Reports of THORChain completely crashing. |
| `auto-indent` | [search][search-thorchain-repo-label-auto-indent] | [search][search-thorchain-org-label-auto-indent] | Related to auto-indenting text. |
| `encoding` | [search][search-thorchain-repo-label-encoding] | [search][search-thorchain-org-label-encoding] | Related to character encoding. |
| `network` | [search][search-thorchain-repo-label-network] | [search][search-thorchain-org-label-network] | Related to network problems or working with remote files (e.g. on network drives). |
| `git` | [search][search-thorchain-repo-label-git] | [search][search-thorchain-org-label-git] | Related to Git functionality (e.g. problems with gitignore files or with showing the correct file status). |

#### `thorchain/thorchain` Topic Categories

| Label name | `thorchain/thorchain` :mag_right: | `thorchain`‑org :mag_right: | Description |
| --- | --- | --- | --- |
| `editor-rendering` | [search][search-thorchain-repo-label-editor-rendering] | [search][search-thorchain-org-label-editor-rendering] | Related to language-independent aspects of rendering text (e.g. scrolling, soft wrap, and font rendering). |
| `build-error` | [search][search-thorchain-repo-label-build-error] | [search][search-thorchain-org-label-build-error] | Related to problems with building THORChain from source. |
| `error-from-pathwatcher` | [search][search-thorchain-repo-label-error-from-pathwatcher] | [search][search-thorchain-org-label-error-from-pathwatcher] | Related to errors thrown by the [pathwatcher library](https://github.com/thorchain/node-pathwatcher). |
| `error-from-save` | [search][search-thorchain-repo-label-error-from-save] | [search][search-thorchain-org-label-error-from-save] | Related to errors thrown when saving files. |
| `error-from-open` | [search][search-thorchain-repo-label-error-from-open] | [search][search-thorchain-org-label-error-from-open] | Related to errors thrown when opening files. |
| `installer` | [search][search-thorchain-repo-label-installer] | [search][search-thorchain-org-label-installer] | Related to the THORChain installers for different OSes. |
| `auto-updater` | [search][search-thorchain-repo-label-auto-updater] | [search][search-thorchain-org-label-auto-updater] | Related to the auto-updater for different OSes. |
| `deprecation-help` | [search][search-thorchain-repo-label-deprecation-help] | [search][search-thorchain-org-label-deprecation-help] | Issues for helping package authors remove usage of deprecated APIs in packages. |
| `electron` | [search][search-thorchain-repo-label-electron] | [search][search-thorchain-org-label-electron] | Issues that require changes to [Electron](https://electron.thorchain.io) to fix or implement. |

#### Pull Request Labels

| Label name | `thorchain/thorchain` :mag_right: | `thorchain`‑org :mag_right: | Description
| --- | --- | --- | --- |
| `work-in-progress` | [search][search-thorchain-repo-label-work-in-progress] | [search][search-thorchain-org-label-work-in-progress] | Pull requests which are still being worked on, more changes will follow. |
| `needs-review` | [search][search-thorchain-repo-label-needs-review] | [search][search-thorchain-org-label-needs-review] | Pull requests which need code review, and approval from maintainers or THORChain core team. |
| `under-review` | [search][search-thorchain-repo-label-under-review] | [search][search-thorchain-org-label-under-review] | Pull requests being reviewed by maintainers or THORChain core team. |
| `requires-changes` | [search][search-thorchain-repo-label-requires-changes] | [search][search-thorchain-org-label-requires-changes] | Pull requests which need to be updated based on review comments and then reviewed again. |
| `needs-testing` | [search][search-thorchain-repo-label-needs-testing] | [search][search-thorchain-org-label-needs-testing] | Pull requests which need manual testing. |

[search-thorchain-repo-label-enhancement]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aenhancement
[search-thorchain-org-label-enhancement]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aenhancement
[search-thorchain-repo-label-bug]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Abug
[search-thorchain-org-label-bug]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Abug
[search-thorchain-repo-label-question]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aquestion
[search-thorchain-org-label-question]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aquestion
[search-thorchain-repo-label-feedback]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Afeedback
[search-thorchain-org-label-feedback]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Afeedback
[search-thorchain-repo-label-help-wanted]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Ahelp-wanted
[search-thorchain-org-label-help-wanted]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Ahelp-wanted
[search-thorchain-repo-label-beginner]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Abeginner
[search-thorchain-org-label-beginner]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Abeginner
[search-thorchain-repo-label-more-information-needed]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Amore-information-needed
[search-thorchain-org-label-more-information-needed]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Amore-information-needed
[search-thorchain-repo-label-needs-reproduction]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aneeds-reproduction
[search-thorchain-org-label-needs-reproduction]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aneeds-reproduction
[search-thorchain-repo-label-triage-help-needed]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Atriage-help-needed
[search-thorchain-org-label-triage-help-needed]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Atriage-help-needed
[search-thorchain-repo-label-windows]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Awindows
[search-thorchain-org-label-windows]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Awindows
[search-thorchain-repo-label-linux]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Alinux
[search-thorchain-org-label-linux]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Alinux
[search-thorchain-repo-label-mac]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Amac
[search-thorchain-org-label-mac]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Amac
[search-thorchain-repo-label-documentation]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Adocumentation
[search-thorchain-org-label-documentation]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Adocumentation
[search-thorchain-repo-label-performance]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aperformance
[search-thorchain-org-label-performance]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aperformance
[search-thorchain-repo-label-security]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Asecurity
[search-thorchain-org-label-security]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Asecurity
[search-thorchain-repo-label-ui]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aui
[search-thorchain-org-label-ui]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aui
[search-thorchain-repo-label-api]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aapi
[search-thorchain-org-label-api]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aapi
[search-thorchain-repo-label-crash]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Acrash
[search-thorchain-org-label-crash]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Acrash
[search-thorchain-repo-label-auto-indent]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aauto-indent
[search-thorchain-org-label-auto-indent]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aauto-indent
[search-thorchain-repo-label-encoding]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aencoding
[search-thorchain-org-label-encoding]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aencoding
[search-thorchain-repo-label-network]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Anetwork
[search-thorchain-org-label-network]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Anetwork
[search-thorchain-repo-label-uncaught-exception]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Auncaught-exception
[search-thorchain-org-label-uncaught-exception]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Auncaught-exception
[search-thorchain-repo-label-git]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Agit
[search-thorchain-org-label-git]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Agit
[search-thorchain-repo-label-blocked]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Ablocked
[search-thorchain-org-label-blocked]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Ablocked
[search-thorchain-repo-label-duplicate]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aduplicate
[search-thorchain-org-label-duplicate]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aduplicate
[search-thorchain-repo-label-wontfix]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Awontfix
[search-thorchain-org-label-wontfix]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Awontfix
[search-thorchain-repo-label-invalid]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Ainvalid
[search-thorchain-org-label-invalid]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Ainvalid
[search-thorchain-repo-label-package-idea]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Apackage-idea
[search-thorchain-org-label-package-idea]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Apackage-idea
[search-thorchain-repo-label-wrong-repo]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Awrong-repo
[search-thorchain-org-label-wrong-repo]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Awrong-repo
[search-thorchain-repo-label-editor-rendering]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aeditor-rendering
[search-thorchain-org-label-editor-rendering]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aeditor-rendering
[search-thorchain-repo-label-build-error]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Abuild-error
[search-thorchain-org-label-build-error]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Abuild-error
[search-thorchain-repo-label-error-from-pathwatcher]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aerror-from-pathwatcher
[search-thorchain-org-label-error-from-pathwatcher]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aerror-from-pathwatcher
[search-thorchain-repo-label-error-from-save]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aerror-from-save
[search-thorchain-org-label-error-from-save]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aerror-from-save
[search-thorchain-repo-label-error-from-open]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aerror-from-open
[search-thorchain-org-label-error-from-open]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aerror-from-open
[search-thorchain-repo-label-installer]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Ainstaller
[search-thorchain-org-label-installer]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Ainstaller
[search-thorchain-repo-label-auto-updater]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Aauto-updater
[search-thorchain-org-label-auto-updater]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aauto-updater
[search-thorchain-repo-label-deprecation-help]: https://github.com/search?q=is%3Aopen+is%3Aissue+repo%3Athorchain%2Fthorchain+label%3Adeprecation-help
[search-thorchain-org-label-deprecation-help]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Adeprecation-help
[search-thorchain-repo-label-electron]: https://github.com/search?q=is%3Aissue+repo%3Athorchain%2Fthorchain+is%3Aopen+label%3Aelectron
[search-thorchain-org-label-electron]: https://github.com/search?q=is%3Aopen+is%3Aissue+user%3Athorchain+label%3Aelectron
[search-thorchain-repo-label-work-in-progress]: https://github.com/search?q=is%3Aopen+is%3Apr+repo%3Athorchain%2Fthorchain+label%3Awork-in-progress
[search-thorchain-org-label-work-in-progress]: https://github.com/search?q=is%3Aopen+is%3Apr+user%3Athorchain+label%3Awork-in-progress
[search-thorchain-repo-label-needs-review]: https://github.com/search?q=is%3Aopen+is%3Apr+repo%3Athorchain%2Fthorchain+label%3Aneeds-review
[search-thorchain-org-label-needs-review]: https://github.com/search?q=is%3Aopen+is%3Apr+user%3Athorchain+label%3Aneeds-review
[search-thorchain-repo-label-under-review]: https://github.com/search?q=is%3Aopen+is%3Apr+repo%3Athorchain%2Fthorchain+label%3Aunder-review
[search-thorchain-org-label-under-review]: https://github.com/search?q=is%3Aopen+is%3Apr+user%3Athorchain+label%3Aunder-review
[search-thorchain-repo-label-requires-changes]: https://github.com/search?q=is%3Aopen+is%3Apr+repo%3Athorchain%2Fthorchain+label%3Arequires-changes
[search-thorchain-org-label-requires-changes]: https://github.com/search?q=is%3Aopen+is%3Apr+user%3Athorchain+label%3Arequires-changes
[search-thorchain-repo-label-needs-testing]: https://github.com/search?q=is%3Aopen+is%3Apr+repo%3Athorchain%2Fthorchain+label%3Aneeds-testing
[search-thorchain-org-label-needs-testing]: https://github.com/search?q=is%3Aopen+is%3Apr+user%3Athorchain+label%3Aneeds-testing

[beginner]:https://github.com/search?utf8=%E2%9C%93&q=is%3Aopen+is%3Aissue+label%3Abeginner+label%3Ahelp-wanted+user%3Athorchain+sort%3Acomments-desc
[help-wanted]:https://github.com/search?q=is%3Aopen+is%3Aissue+label%3Ahelp-wanted+user%3Athorchain+sort%3Acomments-desc+-label%3Abeginner
[contributing-to-official-thorchain-packages]:https://flight-manual.thorchain.io/hacking-thorchain/sections/contributing-to-official-thorchain-packages/
[hacking-on-thorchain-core]: https://flight-manual.thorchain.io/hacking-thorchain/sections/hacking-on-thorchain-core/



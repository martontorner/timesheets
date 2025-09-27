# Contributing

Contributions are welcome, but the goal is to keep the project small and focused. Large or broad feature requests may be declined. Forking the repository is encouraged if you want to experiment or implement more complex features. Examples like `dockerize` can serve as references for more sophisticated tools.

## Questions

For questions, use the [Discussions](https://github.com/your-repo/discussions) feature. Avoid opening issues for general inquiries.

## Issues (Feature Requests and Bugs)

- **Feature Requests:** Clearly describe the problem, your proposed solution, and any alternatives considered. Keep requests scoped and actionable.
- **Bugs:** Include a reproducible example, expected behavior, and observed behavior. Provide any logs or error messages to help diagnose the problem.
- **Formatting:** Use clear titles and labels to help maintainers triage efficiently.

## Coding Rules

- Run all linters and formatters before committing.
- Follow existing code style and structure.
- Include tests for any new functionality.

## Branches

- Mainline branch is `main`.
- Maintain a linear history; use rebase when integrating changes.
- Pull requests must be opened for all changes.
- Example workflow for contributing to your own fork:

  ```bash
  git clone https://github.com/your-username/your-fork.git
  git switch -c feature/my-feature-branch
  # make changes
  git commit -s -m "feat: add new feature"
  git push origin feature/my-feature-branch
  ```

  Then open a pull request from your fork to `main`.

## Commits

### Messages

Commits must follow the [Conventional Commits](https://www.conventionalcommits.org/en/v1.0.0/#specification) specification. CI actions check pull requests for proper formatting. This ensures:

- Automatic changelog generation.
- Clear identification of breaking changes.
- Consistent commit history.

### Sign-off

All commits must be signed off:

```bash
git commit -s -m "type(scope): subject"
```

This certifies that you have the right to submit the code under the repository’s license. See [DCO](https://developercertificate.org/) for details.

## Pull Requests

- Open a pull request against `main`.
- **Merge strategies:**
  - Squash commits for small or medium changes.
  - Rebase and fast-forward for large changes with multiple commits.
- Ensure your branch passes all CI checks before requesting a merge.

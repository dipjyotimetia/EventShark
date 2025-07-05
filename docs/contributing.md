# Contributing to Event Shark

Thank you for considering contributing to Event Shark! We welcome contributions from the community to help improve and enhance the project. This document provides guidelines and best practices for contributing to the project.

## How to Contribute

### Submitting Issues

If you encounter any issues or have suggestions for improvements, please submit an issue on the [GitHub repository](https://github.com/dipjyotimetia/EventShark/issues). When submitting an issue, please provide as much detail as possible, including steps to reproduce the issue, expected behavior, and any relevant logs or screenshots.

### Submitting Pull Requests

We welcome pull requests for bug fixes, new features, and improvements. To submit a pull request, follow these steps:

1. Fork the repository on GitHub.
2. Clone your forked repository to your local machine:
   ```sh
   git clone https://github.com/your-username/EventShark.git
   cd EventShark
   ```
3. Create a new branch for your changes:
   ```sh
   git checkout -b my-feature-branch
   ```
4. Make your changes and commit them with a descriptive commit message:
   ```sh
   git commit -am "Add new feature"
   ```
5. Push your changes to your forked repository:
   ```sh
   git push origin my-feature-branch
   ```
6. Open a pull request on the original repository and provide a detailed description of your changes.

### Coding Standards

To ensure consistency and maintainability of the codebase, please adhere to the following coding standards and best practices:

- Follow the existing code style and conventions.
- Write clear and concise code with appropriate comments.
- Ensure your code is well-tested and includes unit tests where applicable.
- Use meaningful variable and function names.
- Avoid introducing unnecessary dependencies.

### Code of Conduct

We are committed to fostering a welcoming and inclusive community. By participating in this project, you agree to abide by our [Code of Conduct](CODE_OF_CONDUCT.md). Please be respectful and considerate of others in all interactions.

### License

By contributing to Event Shark, you agree that your contributions will be licensed under the [MIT License](LICENSE).

### Testing Before PRs

Please run all tests (`make test`) before submitting a pull request. Ensure your changes do not break existing functionality.

### Code and Schema Generation

If you modify Avro schemas or generated code, run the code and schema generation steps:
- `make code-gen` for Go code generation
- `make schema-gen` for JSON schema generation

Thank you for your contributions and support!

# findtarget

`findtarget` is a Go-based tool that retrieves security programs from Bugcrowd and HackerOne based on a specified YAML configuration. It helps security researchers find targets by automating API requests to these platforms.

## Features
- Fetch security programs from **Bugcrowd** and **HackerOne**.
- Filter results based on **reward type, category, and scope**.
- Configure API requests via a simple **YAML file**.
- Supports **environment variables** for authentication.
- Future support planned for **YesWeHack, Open Bug Bounty, and Immunefi**.

## Installation

```sh
go install github.com/e1l1ya/findtarget/cmd/findtarget@latest
```

## Usage

1. **Prepare a YAML configuration file:**

   ```yaml
   findtarget:
     bugcrowd:
       reward: points
       category: website
       scope: wide
       maxPrograms: 2
     hackerone:
       category: website
       scope: wide
       maxPrograms: 2
   ```

2. **Set up HackerOne credentials** (if using HackerOne):

   Create a `.env` file in the root directory:

   ```sh
   # Hackerone Information
   H1_USERNAME="your_username"
   H1_API_KEY="your_api_key"
   ```

3. **Run the tool:**

   ```sh
   go run cmd/findtarget/findtarget.go -t templates/wide.yaml
   ```

## Roadmap
- [ ] Add support for **YesWeHack**
- [ ] Add support for **Open Bug Bounty**
- [ ] Add support for **Immunefi**
- [ ] Enhance filtering options

## Contributing
Contributions are welcome! Feel free to fork the repo, open issues, or submit pull requests.

## License
This project is licensed under the MIT License.

## Contact
For any questions or suggestions, feel free to reach out on GitHub.

---

Happy hacking! 🐞🔍


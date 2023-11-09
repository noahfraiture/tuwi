# tuwi

> Warning : this project is working but mostly in working. This can be considered as an alpha release. There's known issue listed at the end of this document

A terminal user interface application to chat with AI such as the different GPT models, DALL-E 3, and Bard. This project has no final goal and may evolve in the future if I want to do cool stuff, but it should stay in the same flow.

## Explanation 

This project is written in Go with the framework Bubble Tea https://github.com/charmbracelet/bubbletea 

~~The database is couchDB~~

The database system has recently be re-written to simply handle file on the file system of the user

# Usage

1. You should make a `db` directory that will store the conversations
2. You need to create a file `key` with your openai key
3. Run the project with `go run  /path/to/tuwi`
4. Chat

You can navigate with vim motion as displayed in the help menu. Ctrl-z to go back in navigation and ctrl-s on chat to save the conversation 

## Plans

- The new database management came with difficulties to handle. 
  - [ ] Encrypt the data stored in clear, such as conversation history and OpenAI key, for security and privacy reasons
  - [ ] Implement Peterson's algorithm to handle concurrency and block users from loading the same conversation multiple times. Otherwise, the conversations may become unsynchronized and data may be lost.

- Integration of dall-e 3 :
  - [ ] Add a function to call the API and query the image from the given URL.
  - [ ] Implement a more complex solution to display the image in the terminal, compatible with different terminals.

- [ ] Integrate Bard and more AI models.
- [ ] Restructure the project structure.
- [ ] Ask the user for their key at the first connection.
- [ ] Add the possibility to save system messages.
- [ ] Make the help menu more complete.


# Known issue
- Artifact on AI list from the conversation list

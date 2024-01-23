# Development

In this section, we'll explore how to contribute to JioTV Go and make customizations to suit your needs. Whether you're interested in development using Docker Compose or diving into Go programming, we've got you covered.

## Build from Source

To build JioTV Go from source, you'll need to have [Go](https://golang.org/) installed on your system. If you don't have Go installed, you can download it from [here](https://golang.org/dl/).

### Clone

Let's start by cloning the repository:

```bash
git clone https://github.com/rabilrbl/jiotv_go.git
cd jiotv_go
```

### Build

Now, let's build the project:

```bash
go build . -o build/jiotv_go
```

### Run

Finally, let's run the server:

```bash
./build/jiotv_go [commands]
```

## Docker

1. **Run with Docker Compose:**
   - Run the command:
     ```sh
     docker-compose up
     ```
     This will automatically reload the server when you make changes to the code in `.go` and `.html` files.

2. **Running the Server in Background:**
   - To run the server in the background, use:
     ```sh
     docker-compose up -d
     ```

3. **Stop the Server:**
   - To stop the server, run:
     ```sh
     docker-compose down
     ```

4. **Access JioTV Go:**
   - The server will be listening at [http://localhost:5001](http://localhost:5001).

5. **Set Environment Variables:**
   - You can set environment variables in the `.env` file for customizations.

## Local Development

JioTV Go is powered by [Golang](https://golang.org/), making it an exciting project for developers to explore and contribute to. Here's how you can set up and run the server using Go:

1. **Ensure You Have Golang Installed:**
   - First, make sure you have Golang installed on your system.

2. **Start the Server:**
   - Fire up the server with:
     ```sh
     go run ./cmd/jiotv_go
     ```
     Please note that you'll need to stop and restart the server manually when you make changes to the code.

3. **Enable Debugging and Auto-Reloading:**
   - To enable automatic reloading on template changes in the `views` folder and enable debug logs in the console/terminal, set `JIOTV_DEBUG=true` or [`debug` config value to `true`]().

That's it! You're now all set to explore and contribute to JioTV Go. Happy coding! 🖥️👩‍💻👨‍💻

## Customize the Look with TailwindCSS

At JioTV Go, we use the versatile [TailwindCSS](https://tailwindcss.com/) for styling our project. If you're eager to make some style enhancements, here's how you can do it:

1. **Ensure You Have NodeJS Installed:**
   - Make sure you have NodeJS installed on your system.

2. **Navigate to the `web` Directory:**
   - Open a new terminal window and navigate to the project's root directory. Then, switch to the `web` directory by running:
     ```sh
     cd web
     ```

3. **Install Dependencies:**
   
   - Install the necessary dependencies by running:
     ```sh
     npm install
     ```

4. **Real-Time TailwindCSS Updates:**
   
   - To keep TailwindCSS up to date with your changes in real-time, run the following command:
     ```sh
     npm run watch
     ```

5. **Build Minified CSS:**
   - Once you're satisfied with your style modifications, it's time to build the minified CSS file:
     ```sh
     npm run build
     ```

Now you have the flexibility to customize the look and feel of JioTV Go to match your preferences.

## Let's Make JioTV Go Better Together!

### Report Bugs

Found a pesky bug? No worries! Please help us improve JioTV Go by creating an issue [here](https://github.com/rabilrbl/jiotv_go/issues/new). Be sure to include detailed steps to reproduce the bug, describe the expected behavior, and, if possible, attach screenshots. Your feedback is invaluable!

### Ready to Contribute? Join the Journey!

We wholeheartedly welcome your contributions. If you have ideas, fixes, or enhancements in mind, don't hesitate to create a pull request with your changes. For significant alterations, start by creating an issue to discuss your plans with us. Together, we can make JioTV Go even more incredible.

Thank you for considering contributing to JioTV Go, and happy coding!

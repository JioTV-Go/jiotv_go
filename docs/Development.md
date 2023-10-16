# Dive into Development!

In this section, we'll explore how to contribute to JioTV Go and make customizations to suit your needs. Whether you're interested in development using Docker Compose or diving into Go programming, we've got you covered.

## Using Docker Compose

1. **Clone the Repository:**
   - Start by cloning the JioTV Go repository and navigate to the project's root directory.

2. **Run with Docker Compose:**
   - Run the command:
     ```sh
     docker-compose up
     ```
     This will automatically reload the server when you make changes to the code in `.go` and `.html` files.

3. **Running the Server in Background:**
   - To run the server in the background, use:
     ```sh
     docker-compose up -d
     ```

4. **Stop the Server:**
   - To stop the server, run:
     ```sh
     docker-compose down
     ```

5. **Access JioTV Go:**
   - The server will be listening at `http://localhost:5001`.

6. **Set Environment Variables:**
   - You can set environment variables in the `.env` file for customizations.

## Using Go Natively

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
   - To enable automatic reloading on template changes in the `views` folder and enable debug logs in the console/terminal, set `JIOTV_DEBUG=true`.

That's it! You're now all set to explore and contribute to JioTV Go. Happy coding! 🖥️👩‍💻👨‍💻

## Customize the Look with TailwindCSS

At JioTV Go, we use the versatile [TailwindCSS](https://tailwindcss.com/) for styling our project. If you're eager to make some style enhancements, here's how you can do it:

1. **Ensure You Have NodeJS Installed:**
   - Make sure you have NodeJS installed on your system.

2. **Navigate to the `web` Directory:**
   - Open a new terminal window and navigate to the project's root directory.

3. **Install Dependencies:**
   - Switch to the `web` directory by running:
     ```sh
     cd web
     ```
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

## Building JioTV Go from Source

Building JioTV Go from source is straightforward. Follow these simple steps:

1. **Clone the Repository:**
   - After cloning the repository, navigate to the project's directory.

2. **Build JioTV Go:**
   - Run the following command to build JioTV Go:
     ```sh
     go build ./cmd/jiotv_go -o build/jiotv_go
     ```
   This will create a binary named `jiotv_go` in the `build` directory.

These instructions enable you to build JioTV Go from the source code, giving you the ability to customize and extend the application.

## Let's Make JioTV Go Better Together!

### Report Bugs

Found a pesky bug? No worries! Please help us improve JioTV Go by creating an issue [here](https://github.com/rabilrbl/jiotv_go/issues/new). Be sure to include detailed steps to reproduce the bug, describe the expected behavior, and, if possible, attach screenshots. Your feedback is invaluable!

### Ready to Contribute? Join the Journey!

We wholeheartedly welcome your contributions. If you have ideas, fixes, or enhancements in mind, don't hesitate to create a pull request with your changes. For significant alterations, start by creating an issue to discuss your plans with us. Together, we can make JioTV Go even more incredible.

Thank you for considering contributing to JioTV Go, and happy coding!


# Customize the Look with TailwindCSS

JioTV Go is designed to be flexible and customizable, and one way you can tailor the user interface to your liking is by using [TailwindCSS](https://tailwindcss.com/), a popular utility-first CSS framework. This guide will walk you through how to make style enhancements and adjustments to the appearance of JioTV Go.

## Prerequisites

Before you begin customizing the look with TailwindCSS, ensure that you have the following prerequisites:

1. **NodeJS Installed**: Make sure you have NodeJS installed on your system. If you don't have NodeJS installed, you can download it from the official website.

2. **Access to JioTV Go Project**: You should have access to the JioTV Go project files.

## Steps to Customize with TailwindCSS

Here are the steps to customize the look of JioTV Go with TailwindCSS:

1. **Navigate to the `web` Directory**:
   - Open a terminal window and navigate to the project's root directory.

2. **Switch to the `web` Directory**:
   - Run the following command to navigate to the `web` directory, where the TailwindCSS configuration is located:
     ```sh
     cd web
     ```

3. **Install Dependencies**:
   - Install the necessary dependencies by running the following command:
     ```sh
     npm install
     ```

4. **Real-Time Updates**:
   - To keep TailwindCSS up to date with your changes in real-time, run the following command:
     ```sh
     npm run watch
     ```

5. **Customize Styles**:
   - Now you can start customizing the styles. Open the relevant CSS files and make changes to the classes as needed.

6. **Build Minified CSS**:
   - Once you're satisfied with your style modifications, it's time to build the minified CSS file by running:
     ```sh
     npm run build
     ```

## Example Customizations

Here are some common customizations you can make using TailwindCSS:

- Adjusting color schemes.
- Modifying typography (font size, weight, etc.).
- Tweaking spacing and margins.
- Customizing button styles.
- Enhancing the overall layout.

Feel free to experiment and create a unique look for your JioTV Go instance.

## Enjoy a Customized JioTV Go

Once you've made your desired customizations, you'll be able to enjoy a JioTV Go interface that matches your preferences. TailwindCSS's utility classes make it easy to make changes and experiment with different styles.

If you have any questions or need further guidance on customizing the look with TailwindCSS, don't hesitate to reach out to the JioTV Go community or the project maintainers. Your unique style can help make JioTV Go even more enjoyable for yourself and other users.

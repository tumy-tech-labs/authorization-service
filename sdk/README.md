# Authorization SDK for Node.js

## Overview

This SDK provides a convenient way for Node.js applications to interact with an authorization service. It allows developers to generate authentication tokens and evaluate access policies based on predefined rules.

## Features

- **Token Generation**: Generate authentication tokens using client credentials.
- **Policy Evaluation**: Evaluate access policies based on subject, resource, and action.

## Installation

To install the SDK, you can use npm:

```bash
npm install @yourorganization/authorization-sdk
```

## Usage

### Configuration

Before using the SDK, make sure to set up your environment variables. Create a `.env` file in your project root with the following variables:

```dotenv
SDK_CLIENT_ID=your_client_id_here
SDK_CLIENT_SECRET=your_client_secret_here
```

### Example

Here’s how you can use the SDK in your Node.js application (`index.js`):

```javascript
require('dotenv').config(); // Load environment variables from .env file
const AuthorizationSDK = require('@yourorganization/authorization-sdk');

// Initialize SDK with client credentials
const sdk = new AuthorizationSDK({
  clientId: process.env.SDK_CLIENT_ID,
  clientSecret: process.env.SDK_CLIENT_SECRET,
});

// Example usage to evaluate policy
async function evaluatePolicy() {
  const subject = 'user2';
  const resource = 'file2';
  const action = 'read';

  try {
    // Generate token using SDK method
    const token = await sdk.generateToken();

    // Call SDK method to evaluate policy
    const decision = await sdk.checkAccess(token, subject, resource, action);

    console.log('Policy Evaluation Result:', decision);
  } catch (error) {
    console.error('Error:', error.message);
  }
}

// Call the example function
evaluatePolicy();
```

### SDK Methods

#### `generateToken()`

Generates an authentication token using client credentials.

```javascript
async function generateToken() {
  try {
    const token = await sdk.generateToken();
    console.log('Generated Token:', token);
  } catch (error) {
    console.error('Failed to generate token:', error.message);
  }
}
```

#### `checkAccess(token, subject, resource, action)`

Checks access based on the provided authentication token, subject, resource, and action.

```javascript
async function checkAccess() {
  const token = 'your_generated_token';
  const subject = 'user1';
  const resource = 'file1';
  const action = 'read';

  try {
    const decision = await sdk.checkAccess(token, subject, resource, action);
    console.log('Access Decision:', decision);
  } catch (error) {
    console.error('Failed to check access:', error.message);
  }
}
```

## Contributing

Contributions are welcome! If you find any issues or have suggestions for improvements, please open an issue or submit a pull request on GitHub.

## License

This SDK is licensed under the MIT License. See the LICENSE file for more details.

---

Adjust the placeholders (`@yourorganization/authorization-sdk`) with your actual package name if you plan to publish it on npm. Ensure to provide clear and concise examples, detailed usage instructions, and guidelines for setting up and configuring the SDK in different environments.
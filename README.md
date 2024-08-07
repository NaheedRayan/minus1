


# Installation Steps

1. **Clone the repo**
   ```bash
   cd
   git clone https://github.com/NaheedRayan/minus1.git
   cd minus1
   ```

2. **Obtain Your Gemini API Key**
   
   Get your Gemini API key from [here](https://makersuite.google.com/app/apikey) and configure the `config.json` file.

   ```bash
   nano config.json
   ```

3. **Build the Executable** (optional)
   
   You have to install golang to Build the executable with the following command:

   ```bash
   go build -o minus1
   ```


4. **Update Path**

   Add the directory to your system's PATH:

   ```bash
   echo '## minus1 setup' >> ~/.bashrc
   echo 'export PATH=$PATH:$HOME/minus1' >> ~/.bashrc
   ```

   Apply the changes:

   ```bash
   source ~/.bashrc
   ```

5. **Enjoy**

   Write in terminal

   ```bash
   minus1
   ```



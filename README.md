


# Installation Steps

1. **Obtain Your Gemini API Key**
   
   Get your Gemini API key from [here](https://makersuite.google.com/app/apikey) and configure the `config.json` file.

   ```bash
   nano config.json
   ```

2. **Build the Executable**
   
   Build the executable with the following command:

   ```bash
   go build -o minus1
   ```

3. **Create Directory and Move Files**

   Create a directory named `minus1` in `/usr/local/bin`:

   ```bash
   sudo mkdir -p /usr/local/minus1
   ```

   Move the built executable and configuration file to this directory:

   ```bash
   sudo mv minus1 /usr/local/minus1
   sudo cp config.json /usr/local/minus1
   ```

4. **Create and Set Permissions for Log Files**

   Create log files and set permissions:

   ```bash
   sudo touch /usr/local/minus1/cmdList.txt
   sudo touch /usr/local/minus1/cmdLog.txt

   sudo chmod 777 /usr/local/minus1/cmdList.txt
   sudo chmod 777 /usr/local/minus1/cmdLog.txt
   ```

5. **Update Path**

   Add the directory to your system's PATH:

   ```bash
   echo '## minus1 setup' >> ~/.bashrc
   echo 'export PATH=$PATH:/usr/local/minus1' >> ~/.bashrc
   ```

   Apply the changes:

   ```bash
   source ~/.bashrc
   ```

6. **Enjoy**

   Write in terminal

   ```bash
   minus1
   ```



# Resource Propagation across Multiple Environments

This repository provides instructions on how to continuously propagate resources across multiple environments using the IAM-CTL tool.

## Recommended Workflow

To manage resources when multiple environments exist, follow this recommended workflow:

1. **Set Up WSO2 Identity Server Instances**: Have separate WSO2 Identity Server instances for each environment, hosted somewhere accessible by the public.

2. **Create Management Application**: Create an application with the Management Application enabled in the target Identity Server. Update the OAuth inbound authentication configuration with a dummy callback URL. Make note of the Client ID and Client Secret for later use. *(Automation will be implemented in the future)*

3. **Git Repository Setup**:

   - Create a Git repository with branches dedicated to each environment. Use the environment name (same as the branch name) as an environment variable with the key `ENV` for each branch. This is where the resource config files will reside for that environment.

   - **Manual Approach**:
     - Clone the repository to your local machine.
     - Download the latest release of the IAM-CTL tool from [GitHub Releases](https://github.com/wso2-extensions/identity-tools-cli/releases) and unzip it.
     - Run the following command, replacing `<path to the cloned repo folder>` with the appropriate path: 
       ```
       iamctl setupCLI -d "<path to the cloned repo folder>"
       ```
     - Create subfolders for each environment in the `config` folder and configure each `serverConfig.json` file with the relevant server details and other configurations.
     - *Note: Variables of `serverConfig.json` can also be added as environment values. Refer to the [documentation](<LINK>) for more information on loading server configurations from environment variables.*

   - **Workflow Automation**:
     - Set up the environment variable. Refer to the [documentation](<LINK>) for more information on loading server configurations from environment variables.
     - Add a GitHub Action with the custom action `@action/setup` (replace with the correct action reference).
     - Trigger the action.
     - An automatically generated `config` folder will be added to the repository.

4. **Export Resources**:
   - Export resources from the lower environment to the lower environment branch using another GitHub Action with the custom action `@action/export`.
   - A pull request will be created with the changes. Review and verify the correctness of the propagation, and then merge the pull request.

5. **Environment-specific Variables**:
   - If there are environment-specific variables, add keyword placeholders to the exported files and add the relevant keyword mapping to the tool configurations.

6. **Merge to Higher Environment Branch**:
   - Merge the lower environment branch to the higher environment branch. *(Automation can be implemented to automatically merge the pull request from the lower environment to the higher environment branch.)*

7. **Import Resources**:
   - Use the import command to deploy the resources from the repository to the higher environment using another GitHub Action with the custom action `@action/import`.
   - *(Automation can also be implemented by triggering the Git action automatically upon merging to the branch.)*

8. **Update and Import New Changes**:
   - When new changes are added to the lower environment, export the resources again to update the resource configurations in the branch.
   - Merge the changes to the higher environment repository and import them back to the higher environments.
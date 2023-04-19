# Resource Propagation across Multiple Environments

This section describes how the IAM-CTL tool can be used to continously propagate resources across multiple environments.

### Recommended workflow
Here is the recommended workflow for managing resources using the tool when multiple environments exist:
1. Have separate WSO2 Identity Server instances for each environment.
2. Create a separate config folder for each environment with the relevant server details and other configurations.
3. Export resources from the lower environment to a local directory, which would be the common resource configuration directory for all environments.
4. If there are environment specific variables, add keyword placeholders to the exported files, and add the relevant keyword mapping to the tool configs.
5. Use the import command to deploy the resources from the local directory to higher environments.
6. When new changes are added to the lower environment, export the resources again to update the local resource configurations and import them back to higher environments.


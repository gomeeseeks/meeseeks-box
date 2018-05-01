# Kubernetes deployment

To use this deployment file you will need to:

1. Get a slack token for your bot
2. Run `echo -n $SLACK_TOKEN | base64 > slack-token`
3. Run `kubectl create secret generic slack-token --from-file=./slack-token`
4. Create the configmap with `kubectl create -f meeseeks-configmap.yaml`
5. Create the deployment with `kubectl create -f meeseeks-deployment.yaml`
6. Profit

This deployment will have no persistence for the meeseeks database file, in
case you do need to keep this file around simply use any other storage
technique like NFS or whatever you have available in your cluster.

# The State Store

kOps has the notion of a 'state store'; a location where we store the configuration of your cluster. State is stored
here not only when you first create a cluster, but also you can change the state and apply changes to a running cluster.

Eventually, kubernetes services will also pull from the state store, so that we don't need to marshal all our
configuration through a channel like user-data. (This is currently done for secrets and SSL keys, for example,
though we have to copy the data from the state store to a file where components like kubelet can read them).

The state store uses kOps's VFS implementation, so can in theory be stored anywhere.
As of now the following state stores are supported:

* Amazon AWS S3 (`s3://`)
* local filesystem (`file://`) (only for dry-run purposes, see [note](#local-filesystem-state-stores) below)
* Digital Ocean (`do://`)
* MemFS (memfs://)
* Google Cloud (`gs://`)
* Kubernetes (`k8s://`)
* OpenStack Swift (`swift://`)

The state store is just files; you can copy the files down and put them into git (or your preferred version control system).

## {statestore}/config

One of the most important files in the state store is the top-level config file. This file stores the main
configuration for your cluster (instance types, zones, etc)\

When you run `kops create cluster`, we create a state store entry for you based on the command line options you specify.
For example, when you run with `--node-size=m4.large`, we actually set a line in the configuration
that looks like `NodeMachineType: m4.large`.

The configuration you specify on the command line is actually just a convenient short-cut to
manually editing the configuration.  Options you specify on the command line are merged into the existing
configuration. If you want to configure advanced options, or prefer a text-based configuration, you
may prefer to just edit the config file with `kops edit cluster`.

Because the configuration is merged, this is how you can just specify the changed arguments when
reconfiguring your cluster - for example just `kops create cluster` after a dry-run.

## State store configuration

There are a few ways to configure your state store. In priority order:

+ command line argument `--state s3://yourstatestore`
+ environment variable `export KOPS_STATE_STORE=s3://yourstatestore`
+ config file `$HOME/.kops.yaml`
+ config file `$HOME/.kops/config`

## Local filesystem state stores
{{ kops_feature_table(kops_added_default='1.17') }}

The local filesystem state store (`file://`) is **not** functional for running clusters. It is permitted so as to enable review workflows.

For example, in a review workflow, it can be desirable to check a set of untrusted changes before they are applied to real infrastructure. If submitted untrusted changes to configuration files are naively run by `kops replace`, then kOps would overwrite the state store used by production infrastructure with changes which have not yet been approved. This is dangerous.

Instead, a review workflow may download the contents of the state bucket to a local directory (using `aws s3 sync` or similar), set the state store to the local directory (e.g. `--state file:///path/to/state/store`), and then run `kops replace` and `kops update` (but for a dry-run only - _not_ `kops update --yes`). This allows the review process to make changes to a local copy of the state bucket, and check those changes, without touching the production state bucket or production infrastructure.

Trying to use a local filesystem state store for real (i.e. `kops update --yes`) clusters will not work since the Kubernetes nodes in the cluster need to be able to read from the same state bucket, and the local filesystem will not be mounted to all of the Kubernetes nodes. In theory, a cluster administrator could put the state store on a shared NFS volume that is mounted to the same directory on each of the nodes; however, that use case is not supported as of yet.

### Configuration file example:

`$HOME/.kops/config` might look like this:

```
kops_state_store: s3://yourstatestore
```

## State store variants

### S3 state store

The state store for S3 can be either configured via AWS env variables or directly with custom S3 credentials via env variables. The default for the s3 store is using AWS credentials.

It is possible to set the ACLs for the bucket by setting the env variable `KOPS_STATE_S3_ACL`.

#### AWS S3 config

Normally configured via AWS environment variables or AWS credentials file. The mechanism used to retrieve the credentials is derived from the AWS SDK as follows:

``` golang
config = aws.NewConfig().WithRegion(region)
config = config.WithCredentialsChainVerboseErrors(true)
```

where region is fetched from `AWS_REGION` or from ec2 metadata if we're running within EC2. It defaults to `us-east-1`.

#### Custom s3 compatible store

Your custom s3 state store can be configured by providing S3 environment variables:

- `S3_ENDPOINT`: your custom endpoint
- `S3_REGION`: the region to use
- `S3_ACCESS_KEY_ID`: your access key
- `S3_SECRET_ACCESS_KEY`: your secret key

#### Moving state between S3 buckets

The state store can easily be moved to a different s3 bucket. The steps for a single cluster are as follows:

1. Recursively copy all files from `${OLD_KOPS_STATE_STORE}/${CLUSTER_NAME}` to `${NEW_KOPS_STATE_STORE}/${CLUSTER_NAME}` with `aws s3 sync` or a similar tool.
2. Update the `KOPS_STATE_STORE` environment variable to use the new S3 bucket.
3. Either run `kops edit cluster ${CLUSTER_NAME}` or edit the cluster manifest yaml file. Update `.spec.configBase` to reference the new s3 bucket.
4. Run `kops update cluster ${CLUSTER_NAME} --yes` to apply the changes to the cluster. Newly launched nodes will now retrieve their dependent files from the new S3 bucket. The files in the old bucket are now safe to be deleted.

Repeat for each cluster needing to be moved.

#### Cross Account State-store

Many enterprises prefer to run many AWS accounts. In these setups, having a shared cross-account S3 bucket for state may make inventory and management easier.
Consider the S3 bucket living in Account B and the kOps cluster living in Account A. In order to achieve this, you first need to let Account A access the s3 bucket. This is done by adding the following _bucket policy_ on the S3 bucket:

```json
{
    "Id": "123",
    "Version": "2012-10-17",
    "Statement": [
        {
            "Sid": "123",
            "Action": [
                "s3:*"
            ],
            "Effect": "Allow",
            "Resource": [
                "arn:aws:s3:::<state-store-bucket>",
                "arn:aws:s3:::<state-store-bucket>/*"
            ],
            "Principal": {
                "AWS": [
                    "arn:aws:iam::<account-a>:root"
                ]
            }
        }
    ]
}
```

kOps will then use that bucket as if it was in the remote account, including creating appropriate IAM policies that limits nodes from doing bad things.
Note that any user/role with full S3 access will be able to delete any cluster from the state store, but may not delete any instances or other things outside of S3.

## Digital Ocean (do://)

DigitalOcean storage is configured as a flavor of a S3 store.

## OpenStack Swift (swift://)

The swift store can be configured by providing your OpenStack credentials and configuration in environment variables:

- `OS_AUTH_URL`: the identity endpoint to authenticate against
- `OS_USERNAME`: the username to use
- `OS_USERID`: the user ID
- `OS_PASSWORD`: the password for the useraccount
- `OS_TENANT_ID`: the tenant id
- `OS_TENANT_NAME`: the tenant name
- `OS_PROJECT_ID`: the project id
- `OS_PROJECT_NAME`: the project name
- `OS_DOMAIN_ID`: the domain ID
- `OS_DOMAIN_NAME`: the domain name
- `OS_APPLICATION_CREDENTIAL_ID`: application credential ID
- `OS_APPLICATION_CREDENTIAL_NAME`: application credential name
- `OS_APPLICATION_CREDENTIAL_SECRET`: application secret

The mechanism used to retrieve the credentials is derived from the [gophercloud OpenStack SDK](https://pkg.go.dev/github.com/gophercloud/gophercloud).

A credentials file with `OPENSTACK_CREDENTIAL_FILE` or a config derived from your personal credentials living in `$HOME/.openstack/config` can also be used to configure your store.

## Google Cloud (gs://)

The state store config for google cloud will be derived by the google storage client SDK as follows:

``` golang
scope := storage.DevstorageReadWriteScope

httpClient, err := google.DefaultClient(context.Background(), scope)
if err != nil {
	return nil, fmt.Errorf("error building GCS HTTP client: %v", err)
}

gcsClient, err := storage.New(httpClient)

```

## Vault (vault://)
{{ kops_feature_table(kops_added_ff='1.19') }}

kOps has support for using Vault as state store. It is currently an experimental feature and you have to enable the `VFSVaultSupport` feature flag to enable it.

The goal of the vault store is to be a safe storage for the kOps keys and secrets store. It will not work to use this as a kOps registry/config store. Among other things, etcd-manager is unable to read VFS control files from vault. Vault also cannot be used as backend for etcd backups.


```sh
export KOPS_FEATURE_FLAGS=VFSVaultSupport
```

### Node authentication and configuration
The vault store uses IAM auth to authenticate against the vault server and expects the vault auth plugin to be mounted on `/aws`.

Instructions for configuring your vault server to accept IAM authentication are at https://learn.hashicorp.com/vault/identity-access-management/iam-authentication

To configure kOps to use the Vault store, add this to the cluster spec:

```yaml
spec:
  secretStore: vault://<vault>:<port>/<kv2 mount>/clusters/<clustername>/secrets
  keyStore: vault://<vault>:<port>/<kv2 mount>/clusters/<clustername>/keys
```

Each of the paths specified above can be configurable, but they must be unique across all clusters. You can also not use the same path as both `stateStore` and `keyStore`.

After launching your cluster you need to add the cluster roles to Vault, binding them to the cluster's IAM identity and granting them access to the appropriate secrets and keys. The nodes will wait until they can authenticate before completing provisioning.

#### Vault policies
Note that contrary to the S3 state store, kOps will not provision any policies for you. You have to provide roles for both operators and nodes.

Using the example paths above, a policy for the cluster nodes can be:

```
path "kv-v2/metadata/<clustername>" {
  capabilities = ["list"]
}

path "kv-v2/metadata/clusters/<clustername>/*" {
  capabilities = ["list", "read"]
}

path "kv-v2/data/clusters/<clustername>/*" {
  capabilities = ["read"]
}
```

Once you add this policy, you can assign it to the IAM roles like this:

```sh
vault write auth/aws/role/masters.<clustername> auth_type=iam \
              bound_iam_principal_arn=arn:aws:iam::<account>:role/masters.<clustername> policies=<policy> max_ttl=500h
vault write auth/aws/role/nodes.<clustername> auth_type=iam \
              bound_iam_principal_arn=arn:aws:iam::<account>:role/nodes.<clustername> policies=<policy> max_ttl=500h
vault write auth/aws/config/client iam_server_id_header_value="<vault server hostname>"
```

Note that if you re-provision your cluster, you need to re-run the above in order for Vault to update the role internal IDs.

Vault will use TLS by default. If you want to use plaintext instead, add `?tls=false` to the url.

### Client configuration

The `kops` CLI only expects the `VAULT_TOKEN` environment variable to be set to a valid token. You can use any authentication method to obtain a token and then set it manually if the authentication method does not do that automatically.

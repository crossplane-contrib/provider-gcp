# Authenticating to Google Cloud APIs

`provider-gcp` requires credentials to be provided in order to authenticate to
the Google Cloud APIs. This can be done in one of the following ways:

- Authenticating using a base-64 encoded service account key in a Kubernetes
  `Secret`. This is described in detail [here](https://crossplane.io/docs/v1.6/getting-started/install-configure.html#get-gcp-account-keyfile).
- Authenticating using [Workload Identity](https://cloud.google.com/kubernetes-engine/docs/concepts/workload-identity).
  This is described in the [section below](#authenticating-with-workload-identity).

## Authenticating with Workload Identity

*Note: This method is supported in `provider-gcp` v0.20.0 and later.*

Using Workload Identity requires some additional setup.
Many of the steps can also be found in the [documentation](https://cloud.google.com/kubernetes-engine/docs/how-to/workload-identity).

### Steps

These steps assume you already have a running GKE cluster which has already
enabled Workload Identity and has a sufficiently large node pool.

#### 0. Prepare your variables

In the following sections, you'll need to name your resources.
Define the variables below with any names valid in Kubernetes or GCP so that you
can smoothly set it up:

```console
$ PROJECT_ID=<YOUR_GCP_PROJECT_ID>                               # e.g.) acme-prod
$ PROVIDER_GCP=<YOUR_PROVIDER_GCP_NAME>                          # e.g.) provider-gcp
$ VERSION=<YOUR_PROVIDER_GCP_VERSION>                            # e.g.) 0.20.0
$ GCP_SERVICE_ACCOUNT=<YOUR_CROSSPLANE_GCP_SERVICE_ACCOUNT_NAME> # e.g.) crossplane
$ ROLE=<YOUR_ROLE_FOR_CROSSPLANE_GCP_SERVICE_ACCOUNT>            # e.g.) roles/cloudsql.admin
$ CONTROLLER_CONFIG=<YOUR_CONTROLLER_CONFIG_NAME>                # e.g.) gcp-config (Optional)
```

#### 1. Install Crossplane

Install Crossplane from `stable` channel:

```bash
$ helm repo add crossplane-stable https://charts.crossplane.io/stable
$ helm install crossplane --create-namespace --namespace crossplane-system crossplane-stable/crossplane
```

`provider-gcp` can be installed with either the [Crossplane CLI](https://crossplane.io/docs/v1.6/getting-started/install-configure.html#install-crossplane-cli)
or a `Provider` resource as below:

```console
$ cat <<EOF | kubectl apply -f -
apiVersion: pkg.crossplane.io/v1
kind: Provider
metadata:
  name: ${PROVIDER_GCP}
spec:
  package: crossplane/provider-gcp:v${VERSION} # v0.20.0 or later
  controllerConfigRef:
    name: ${CONTROLLER_CONFIG}
EOF
```

#### 2. Configure service accounts to use Workload Identity

Create a GCP service account, which will be used for provisioning actual
infrastructure in GCP, and grant IAM roles you need for accessing the Google
Cloud APIs:

```console
$ gcloud iam service-accounts create ${GCP_SERVICE_ACCOUNT}
$ gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member "serviceAccount:${GCP_SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role ${ROLE} \
    --project ${PROJECT_ID}
```

Get the name of your current `ProviderRevision` of this provider:

```console
$ REVISION=$(kubectl get providers.pkg.crossplane.io ${PROVIDER_GCP} -o jsonpath="{.status.currentRevision}")
```

Next, you'll configure IAM to use Workload Identity.
In this step, you can choose one of the following options to configure service accounts:

- [Option 1] Use a Kubernetes `ServiceAccount` managed by a provider's controller.
- [Option 2] Use a Kubernetes `ServiceAccount` which you created and is specified to `.spec.serviceAccountName`
  in a [`ControllerConfig`](https://doc.crds.dev/github.com/crossplane/crossplane/pkg.crossplane.io/ControllerConfig/v1alpha1@v1.6.2).

##### 2.1. [Option 1] Use a controller-managed `ServiceAccount`

Specify a Kubernetes `ServiceAccount` with the revision you got in the last
step:

```console
$ KUBERNETES_SERVICE_ACCOUNT=${REVISION}
```

##### 2.1. [Option 2] Use a user-managed `ServiceAccount`

Name your Kubernetes `ServiceAccount`:

```console
$ KUBERNETES_SERVICE_ACCOUNT=<YOUR_KUBERNETES_SERVICE_ACCOUNT>
```

Create a `ServiceAccount`, `ControllerConfig`, and `ClusterRoleBinding`:

```console
$ cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name: ${KUBERNETES_SERVICE_ACCOUNT}
  namespace: crossplane-system
---
apiVersion: pkg.crossplane.io/v1alpha1
kind: ControllerConfig
metadata:
  name: ${CONTROLLER_CONFIG}
spec:
  serviceAccountName: ${KUBERNETES_SERVICE_ACCOUNT}
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: crossplane:provider:${PROVIDER_GCP}:system
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: crossplane:provider:${REVISION}:system
subjects:
- kind: ServiceAccount
  name: ${KUBERNETES_SERVICE_ACCOUNT}
  namespace: crossplane-system
EOF
```

#### 2.2. Allow the Kubernetes `ServiceAccount` to impersonate the GCP service account

Grant `roles/iam.workloadIdentityUser` to the GCP service account:

```console
$ gcloud iam service-accounts add-iam-policy-binding \
    ${GCP_SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:${PROJECT_ID}.svc.id.goog[crossplane-system/${KUBERNETES_SERVICE_ACCOUNT}]" \
    --project ${PROJECT_ID}
```

Annotate the `ServiceAccount` with the email address of the GCP service account:

```console
$ kubectl annotate serviceaccount ${KUBERNETES_SERVICE_ACCOUNT} \
    iam.gke.io/gcp-service-account=${GCP_SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com \
    -n crossplane-system
```

### 3. Configure a `ProviderConfig`

Create a `ProviderConfig` with `InjectedIdentity` in `.spec.credentials.source`:

```console
$ cat <<EOF | kubectl apply -f -
apiVersion: gcp.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: default
spec:
  projectID: ${PROJECT_ID}
  credentials:
    source: InjectedIdentity
EOF
```

### 4. Next steps

Now that you have configured `provider-gcp` with Workload Identity supported,
you can [provision infrastructure](https://crossplane.io/docs/v1.6/getting-started/provision-infrastructure).


## Authenticating with Access Tokens

Using temporary Access Tokens will require a process to regenerate an access token before it expires. Luckily we can use a Kubernetes CronJob to fulfill that.

**DISCLAIMER**

*The following method will only work if running the provider in a GKE cluster on GCP. This is because the creation of access tokens requires a service account with Workload Identity enabled.*

### Steps

#### 0. Prepare your variables

In the following sections, you'll need to name your resources.
Define the variables below with any names valid in Kubernetes or GCP so that you
can smoothly set it up:

```console
$ PROJECT_ID=<YOUR_GCP_PROJECT_ID>                               # e.g.) acme-prod
$ GCP_SERVICE_ACCOUNT=<YOUR_CROSSPLANE_GCP_SERVICE_ACCOUNT_NAME> # e.g.) crossplane
$ ROLE=<YOUR_ROLE_FOR_CROSSPLANE_GCP_SERVICE_ACCOUNT>            # e.g.) roles/editor
$ KUBERNETES_SERVICE_ACCOUNT=<YOUR_KUBERNETES_SERVICE_ACCOUNT>   # e.g.) token-generator
$ NAMESPACE=<YOUR_KUBERNETES_NAMESPACE>                          # e.g.) default
$ SECRET_NAME=<YOUR_CREDENTIALS_SECRET_NAME>                     # e.g.) gcp-credentials
$ SECRET_KEY=<NAME_OF_KEY_IN_SECRET>                             # e.g.) token
```

#### 1. Configure service accounts to use Workload Identity

Create a GCP service account, which will be used for provisioning actual
infrastructure in GCP, and grant IAM roles you need for accessing the Google
Cloud APIs:

```console
$ gcloud iam service-accounts create ${GCP_SERVICE_ACCOUNT} --project ${PROJECT_ID}
$ gcloud projects add-iam-policy-binding ${PROJECT_ID} \
    --member "serviceAccount:${GCP_SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com" \
    --role ${ROLE} \
    --project ${PROJECT_ID}
```

#### 2. Create resources to generate an access-token
Create the Kubernetes service account, RBAC, and CronJob to generate the temporary access-token
```console
$ cat <<EOF | kubectl apply -f -
apiVersion: v1
kind: ServiceAccount
metadata:
  name:  ${KUBERNETES_SERVICE_ACCOUNT}
  namespace: ${NAMESPACE}
---
kind: Role
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: ${KUBERNETES_SERVICE_ACCOUNT}-sync
  namespace: ${NAMESPACE}
rules:
- apiGroups: [""]
  resources:
  - secrets
  verbs:
  - get
  - create
  - patch
---
kind: RoleBinding
apiVersion: rbac.authorization.k8s.io/v1
metadata:
  name: ${KUBERNETES_SERVICE_ACCOUNT}-sync-rb
  namespace: ${NAMESPACE}
subjects:
- kind: ServiceAccount
  name: ${KUBERNETES_SERVICE_ACCOUNT}
roleRef:
  kind: Role
  name: ${KUBERNETES_SERVICE_ACCOUNT}-sync
  apiGroup: ""
---
apiVersion: batch/v1
kind: CronJob
metadata:
  name: ${KUBERNETES_SERVICE_ACCOUNT}-credentials-sync
  namespace: ${NAMESPACE}
spec:
  suspend: false
  schedule: "*/45 * * * *"
  failedJobsHistoryLimit: 1
  successfulJobsHistoryLimit: 1
  concurrencyPolicy: Forbid
  startingDeadlineSeconds: 1800
  jobTemplate:
    spec:
      template:
        spec:
          serviceAccountName:  ${KUBERNETES_SERVICE_ACCOUNT}
          restartPolicy: Never
          containers:
            - image: google/cloud-sdk:debian_component_based
              name: create-access-token
              imagePullPolicy: IfNotPresent
              livenessProbe:
                exec:
                  command:
                  - gcloud
                  - version
              readinessProbe:
                exec:
                  command:
                  - gcloud
                  - version
              env:
                - name: SECRET_NAME
                  value: ${SECRET_NAME}
                - name: SECRET_KEY
                  value: ${SECRET_KEY}
              command:
                - /bin/bash
                - -ce
                - |-
                  kubectl create secret generic $SECRET_NAME \
                    --dry-run=client \
                    --from-literal=$SECRET_KEY=\$(gcloud auth print-access-token) \
                    -o yaml | kubectl apply -f -
              resources:
                requests:
                  cpu: 250m
                  memory: 256Mi
                limits:
                  cpu: 500m
                  memory: 512Mi
EOF
```
Grant `roles/iam.workloadIdentityUser` to the GCP service account:

```console
$ gcloud iam service-accounts add-iam-policy-binding \
    ${GCP_SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com \
    --role roles/iam.workloadIdentityUser \
    --member "serviceAccount:${PROJECT_ID}.svc.id.goog[${NAMESPACE}/${KUBERNETES_SERVICE_ACCOUNT}]" \
    --project ${PROJECT_ID}
```

Annotate the `ServiceAccount` with the email address of the GCP service account:

```console
$ kubectl annotate serviceaccount ${KUBERNETES_SERVICE_ACCOUNT} \
    iam.gke.io/gcp-service-account=${GCP_SERVICE_ACCOUNT}@${PROJECT_ID}.iam.gserviceaccount.com \
    -n ${NAMESPACE}
```

#### 3. Create initial Access Token
```console
kubectl -n ${NAMESPACE} create job --from=cronjob/${KUBERNETES_SERVICE_ACCOUNT}-credentials-sync cred-sync-001
```

#### 4. Create ProviderConfig
```console
$ cat <<EOF | kubectl apply -f -
apiVersion: gcp.crossplane.io/v1beta1
kind: ProviderConfig
metadata:
  name: gcp-provider
spec:
  projectID: ${PROJECT_ID}
  credentials:
    source: Secret
    secretRef:
      name: ${SECRET_NAME}
      namespace: ${NAMESPACE}
      key: ${SECRET_KEY}
EOF

# Upgrading to v0.18.x

Starting with `v0.18.0` version, [provider-gcp] will use `v1` version of GCP
APIs instead of `v1beta1` for all resource. Please note that, this is not
about Kubernetes API Versions for managed resource types, rather about the 
version of Google Cloud APIs that our controllers use. Users relying on beta 
features suggested to use [provider-gcp-beta] instead, which could be installed
side-by-side with [provider-gcp]. For more details on this, please see
[RFC issue for provider-gcp-beta] and description of [the PR switching v1].

This resulted in breaking API changes in `v0.18.x` for `GKECluster` and
`NodePool` types which require the manual migration steps in this document.
Please note, even if you don't have these resources, you should still need to
follow the guide for a successful provider package upgrade, where unrelated
steps would be a no-op. 

Please note, if you have already attempted an upgrade without following the
migration steps below, you could end up with an Unhealthy `ProviderRevision`
with the following event:

```
Warning  SyncPackage        21s (x3 over 53s)  packages/providerrevision.pkg.crossplane.io  cannot establish control of object: CustomResourceDefinition.apiextensions.k8s.io "nodepools.container.gcp.crossplane.io" is invalid: status.storedVersions[0]: Invalid value: "v1alpha1": must appear in spec.versions
```

In this state, you could either revert the upgrade (by just setting 
`spec.package` in provider package CR back to the previous version) or have a
successful upgrade by still following all migration steps skipping the 6th one
(i.e. upgrade your provider package).

## Migration Steps

1. Stop the Crossplane controllers by scaling down Crossplane deployment to 0
   replicas. This is to ensure that our manual steps throughout this document
   not to get overridden by any existing composite resource composing resources
   with the objects that we are interacting here.

   ```
   kubectl -n crossplane-system scale deployment crossplane --replicas=0
   ```

2. Export and store your `NodePool` and `GKECluster` resources in a local file
   as a reference for importing after upgrade.

   ```
   kubectl get nodepools.container.gcp.crossplane.io -o yaml > my-nodepools.yaml
   kubectl get gkeclusters.container.gcp.crossplane.io -o yaml > my-gkeclusters.yaml
   ```

   Verify the contents of the files, before moving further.


4. Mark your `NodePool` and `GKECluster` resources for `deletionPolicy` `Orphan`
   to make sure that they are not deleted once custom resource objects in the
   cluster deleted.

   ```
   kubectl get nodepools.container.gcp.crossplane.io -o name | xargs -n1 kubectl patch  -p '{"spec": {"deletionPolicy": "Orphan"}}' --type merge
   kubectl get gkeclusters.container.gcp.crossplane.io -o name | xargs -n1 kubectl patch  -p '{"spec": {"deletionPolicy": "Orphan"}}' --type merge
   ```

5. Delete orphaned objects of type `NodePool` and `GKECluster`.

   ```
   kubectl delete nodepools.container.gcp.crossplane.io --all
   kubectl delete gkeclusters.container.gcp.crossplane.io --all
   ```
   
6. Delete the deprecated CRD types.
   ```
   kubectl delete crd nodepools.container.gcp.crossplane.io
   kubectl delete crd gkeclusters.container.gcp.crossplane.io
   ```


7. Upgrade your provider package to `v0.18.x` by changing `spec.package` in 
   `provider.pkg.crossplane.io` resource.


8. Start the Crossplane controllers by scaling Crossplane deployment back
   to 1 replicas.

   ```
   kubectl -n crossplane-system scale deployment crossplane --replicas=1
   ```

9. Wait until provider package to be `INSTALLED` and `HEALTHY`.

   ```
   kubectl get provider.pkg.crossplane.io
   ```

10. Import existing `NodePools` and `Clusters` by following [the importing existing resources section below](#importing-existing-nodepools-and-clusters).

11. If you have Compositions with `NodePools` and/or `Clusters`, update your
    compositions by following [the updating compositions section below](#updating-compositions).

### Importing existing NodePools and Clusters

At this point, you would need to [import] your existing `NodePool` and
`GKECluster` (which renamed as `Cluster`) resources using the new schema.
See the API changes for each resource in the description of the
[the PR switching v1].

Now, you can either use `NodePool` and `Cluster` types in [provider-gcp] or 
[provider-gcp-beta] depending on your needs (i.e. relying on a beta feature or
not).

#### Importing as Stable

To import your resources for stable types, use the following templates by
filling required fields using the information from `my-nodepools.yaml` and
`my-gkeclusters.yaml` files that we created in the beginning of migration steps.

**Template to import `Cluster`:**

```
apiVersion: container.gcp.crossplane.io/v1beta2
kind: Cluster
metadata:
  name: <to-be-filled>
  annotations:
    crossplane.io/external-name: <to-be-filled>
spec:
  # Consider setting deletionPolicy as "Orphan" to make sure you don't delete
  # the cloud resource until you successfully validate the import
  # deletionPolicy: Orphan
  forProvider:
    location: <to-be-filled>
  providerConfigRef:
    name: <to-be-filled>
```

**Template to import `NodePool`:**

```
apiVersion: container.gcp.crossplane.io/v1beta1
kind: NodePool
metadata:
  name: <to-be-filled>
  annotations:
    crossplane.io/external-name: <to-be-filled>
spec:
  # Consider setting deletionPolicy as "Orphan" to make sure you don't delete
  # the cloud resource until you successfully validate the import
  # deletionPolicy: Orphan
  forProvider:
    cluster: <to-be-filled>
  providerConfigRef:
    name: <to-be-filled>
```

##### Steps

1. Create resources using the above manifests
2. Verify that resources are `READY` and `SYNCED`

#### Importing as Beta

**Template to import `Cluster`:**

```
apiVersion: container.beta.gcp.crossplane.io/v1beta1
kind: Cluster
metadata:
  name: <to-be-filled>
  annotations:
    crossplane.io/external-name: <to-be-filled>
spec:
  # Consider setting deletionPolicy as "Orphan" to make sure you don't delete
  # the cloud resource until you successfully validate the import
  # deletionPolicy: Orphan
  forProvider:
    location: <to-be-filled>
  providerConfigRef:
    name: <to-be-filled>
```

**Template to import `NodePool`:**

```
apiVersion: container.beta.gcp.crossplane.io/v1alpha1
kind: NodePool
metadata:
  name: <to-be-filled>
  annotations:
    crossplane.io/external-name: <to-be-filled>
spec:
  # Consider setting deletionPolicy as "Orphan" to make sure you don't delete
  # the cloud resource until you successfully validate the import
  # deletionPolicy: Orphan
  forProvider:
    cluster: <to-be-filled>
  providerConfigRef:
    name: <to-be-filled>
```

##### Steps

1. Deploy [provider-gcp-beta] on the cluster.
2. Create `ProviderConfig` resource for [provider-gcp-beta].
3. Create resources using the above manifests.
4. Verify that resources are `READY` and `SYNCED`.

### Updating compositions

If you have compositions relying on the deprecated types, you would need to
update them as well to use the new types with proper fields and configuration
depending on whether you have chosen `beta` or `stable`. Until you switch
using the new types, your composite resources would fail since the old types no
longer exist.

To prevent the updates in the compositions to trigger creation of new composed
resources, please follow the steps below:

1. Stop the Crossplane controllers by scaling down Crossplane deployment to 0
   replicas.

   ```
   kubectl -n crossplane-system scale deployment crossplane --replicas=0
   ```

2. Update your compositions to use new types and configurations. If you have
   installed via a configuration package, update package to the new version
   with changes.

3. Update `resourceRefs` in the spec of your composite resources to point
   the imported resources, for example:

   ```
   spec:
     compositionRef:
       name: gke.gcp.platformref.crossplane.io
     compositionUpdatePolicy: Automatic
     id: platform-ref-gcp-cluster
     parameters:
       networkRef:
         id: platform-ref-gcp-network
       nodes:
         count: 3
         size: small
     resourceRefs:
     - apiVersion: container.gcp.crossplane.io/v1beta2 # <--- make sure the apiVersion matches with which you've used during import (e.g. beta vs stable)
       kind: Cluster # <--- make sure type is correct
       name: platform-ref-gcp-cluster-mwx8t-5j9hv # <--- make sure this is the name of the imported cluster
     - apiVersion: container.gcp.crossplane.io/v1beta1 # <--- make sure the apiVersion matches with which you've used during import
       kind: NodePool # <--- make sure type is correct
       name: platform-ref-gcp-cluster-mwx8t-klb7w # <--- make sure this is the name of the imported nodepool
     - apiVersion: helm.crossplane.io/v1beta1
       kind: ProviderConfig
       name: platform-ref-gcp-cluster
     writeConnectionSecretToRef:
     ...
   ```

4. Start the Crossplane controllers by scaling Crossplane deployment back
   to 1 replicas.

   ```
   kubectl -n crossplane-system scale deployment crossplane --replicas=1
   ```

[provider-gcp]: https://github.com/crossplane/provider-gcp
[provider-gcp-beta]: https://github.com/crossplane/provider-gcp-beta
[RFC issue for provider-gcp-beta]: https://github.com/crossplane/provider-gcp/issues/309
[the PR switching v1]: https://github.com/crossplane/provider-gcp/pull/308
[import]: https://crossplane.io/docs/v1.4/concepts/managed-resources.html#importing-existing-resources
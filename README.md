# CR Based APIs: Is It the Right Approach for Your Application?
This is [Dave Smith-Uchida](https://github.com/dsu-igeek)'s and my Cloud Native Rejekts (NA 22) talk, to be presented in Detroit in the Fall of 2022.

## Abstract

In a microservice application, services need to make API calls to one another. Many Kubernetes applications have begun using Custom Resources (CRs) for their APIs.

This approach offers many advantages over REST. CRs are declarative in nature, so such APIs are simple to develop and evolve. Controllers for CR based APIs are easier to scale out than REST based API servers. CR APIs are more secure to boot, since they leverage native Kubernetes security features.

However, there is a cost to these benefits, chiefly that CRs incur an overhead that may not be acceptable for some applications. How can we decide when it is appropriate to use them?

In this talk we explore this mechanism, and go over its advantages and disadvantages versus REST. We will demo a real-life example of a CR based API at work, and measure its performance relative to REST using the open-source tool Kubestr. Finally we will go over some guidelines for deciding whether moving to a CR based API is the right choice for you.

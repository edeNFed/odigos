import { NextApiRequest, NextApiResponse } from "next";
import * as k8s from "@kubernetes/client-node";
import { Socket } from "dgram";
import Vendors, { ObservabilityVendor, VendorObjects } from "@/vendors/index";

async function UpdateDest(req: NextApiRequest, res: NextApiResponse) {
  try {
    const vendor = Vendors.find(
      (v: ObservabilityVendor) => v.name === req.body.destType
    );

    if (!vendor) {
      return res.status(400).json({
        error: `Vendor ${req.body.type} not found`,
      });
    }
    const kubeObjects: VendorObjects = vendor.toObjects(req);

    const kc = new k8s.KubeConfig();
    kc.loadFromDefault();
    const k8sApi = kc.makeApiClient(k8s.CustomObjectsApi);
    const current = await k8sApi.getNamespacedCustomObject(
      "odigos.io",
      "v1alpha1",
      process.env.CURRENT_NS || "odigos-system",
      "destinations",
      req.query.destname as string
    );

    const updated = current.body;
    const { spec }: any = updated;

    if (kubeObjects.Data) {
      spec.data = kubeObjects.Data;
    }

    const resp = await k8sApi.replaceNamespacedCustomObject(
      "odigos.io",
      "v1alpha1",
      process.env.CURRENT_NS || "odigos-system",
      "destinations",
      req.query.destname as string,
      {
        ...updated,
        spec,
      }
    );

    return res.status(200).json({ success: true });
  } catch (err) {
    console.error(`Error updating destination ${req.query.destname}`, err);
    return res.status(500).json({ success: false });
  }
}

async function DeleteDest(req: NextApiRequest, res: NextApiResponse) {
  console.log(`Deleting destination ${req.query.destname}`);
  const kc = new k8s.KubeConfig();
  kc.loadFromDefault();
  const k8sApi = kc.makeApiClient(k8s.CustomObjectsApi);
  await k8sApi.deleteNamespacedCustomObject(
    "odigos.io",
    "v1alpha1",
    process.env.CURRENT_NS || "odigos-system",
    "destinations",
    req.query.destname as string
  );
  return res.status(200).json({ success: true });
}

export default async function handler(
  req: NextApiRequest,
  res: NextApiResponse
) {
  if (req.method === "POST") {
    return UpdateDest(req, res);
  } else if (req.method === "DELETE") {
    return DeleteDest(req, res);
  }

  return res.status(405).end(`Method ${req.method} Not Allowed`);
}

import {
  ObservabilitySignals,
  ObservabilityVendor,
  VendorObjects,
} from "@/vendors/index";
import HoneycombLogo from "@/img/vendor/honeycomb.svg";
import { NextApiRequest } from "next";

export class Honeycomb implements ObservabilityVendor {
  name = "honeycomb";
  displayName = "Honeycomb";
  supportedSignals = [ObservabilitySignals.Traces];

  getLogo = (props: any) => {
    return <HoneycombLogo {...props} />;
  };

  getFields = () => {
    return [
      {
        displayName: "API Key",
        id: "apikey",
        name: "apikey",
        type: "password",
      },
    ];
  };

  toObjects = (req: NextApiRequest) => {
    return {
      Secret: {
        HONEYCOMB_API_KEY: Buffer.from(req.body.apikey).toString("base64"),
      },
    };
  };

  mapDataToFields = (data: any) => {
    return {};
  };
}

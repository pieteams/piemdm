import { describe, it, expect } from "vitest";
import { NodeHelper } from "../src/lib/utils/node-helper.js";

describe("UUID Utils", () => {
  it("generates valid UUID format from NodeHelper", () => {
    const uuid = NodeHelper.generateNodeCode();
    expect(uuid).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
    );
    expect(uuid).toHaveLength(36);
  });

  it("generates unique UUIDs from NodeHelper", () => {
    const uuid1 = NodeHelper.generateNodeCode();
    const uuid2 = NodeHelper.generateNodeCode();
    expect(uuid1).not.toBe(uuid2);
  });

  it("creates workflow nodes with valid UUID from NodeHelper", () => {
    const node = {
      NodeCode: NodeHelper.generateNodeCode(),
      NodeName: "Test Node",
      NodeType: "APPROVAL",
    };

    expect(node.NodeCode).toBeDefined();
    expect(node.NodeCode).toMatch(
      /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i,
    );
    expect(node.NodeName).toBe("Test Node");
    expect(node.NodeType).toBe("APPROVAL");
  });
});

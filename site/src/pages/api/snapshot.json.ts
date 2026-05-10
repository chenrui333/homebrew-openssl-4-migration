import snapshot from "../../../../data/site_snapshot.json";

export const prerender = true;

export function GET() {
  return new Response(JSON.stringify(snapshot, null, 2) + "\n", {
    headers: {
      "content-type": "application/json; charset=utf-8",
    },
  });
}

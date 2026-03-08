# managed by infra/worktree-deployer
# slug: __SLUG__
http:
  routers:
    worktree-__SLUG__:
      rule: "Host(`__PUBLIC_HOST__`) && PathPrefix(`/__PUBLIC_PATH__`)"
      service: worktree-__SLUG__
      middlewares:
        - worktree-__SLUG__-stripprefix
      priority: 60
      tls:
        certResolver: le

  middlewares:
    worktree-__SLUG__-stripprefix:
      stripPrefix:
        prefixes:
          - "/__PUBLIC_PATH__"

  services:
    worktree-__SLUG__:
      loadBalancer:
        servers:
          - url: "__TARGET_URL__"

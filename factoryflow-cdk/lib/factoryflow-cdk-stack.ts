import * as cdk from "aws-cdk-lib";
import * as ecs from "aws-cdk-lib/aws-ecs";
import * as ecsPatterns from "aws-cdk-lib/aws-ecs-patterns";
import * as ec2 from "aws-cdk-lib/aws-ec2";
import * as rds from "aws-cdk-lib/aws-rds";

export class FactoryflowCdkStack extends cdk.Stack {
  constructor(scope: cdk.App, id: string, props?: cdk.StackProps) {
    super(scope, id, props);

    // Create VPC
    const vpc = new ec2.Vpc(this, "FactoryFlowVPC", {
      maxAzs: 2,
    });

    // Create RDS Database
    const database = new rds.DatabaseInstance(this, "FactoryFlowDB", {
      engine: rds.DatabaseInstanceEngine.postgres({
        version: rds.PostgresEngineVersion.VER_15,
      }),
      instanceType: ec2.InstanceType.of(
        ec2.InstanceClass.T3,
        ec2.InstanceSize.MICRO
      ),
      vpc,
      databaseName: "factoryflow",
      credentials: rds.Credentials.fromGeneratedSecret("factoryuser"),
      removalPolicy: cdk.RemovalPolicy.DESTROY,
    });

    // Create ECS Cluster
    const cluster = new ecs.Cluster(this, "FactoryFlowCluster", {
      vpc,
    });

    // Frontend Service
    const frontend = new ecsPatterns.ApplicationLoadBalancedFargateService(
      this,
      "Frontend",
      {
        cluster,
        taskImageOptions: {
          image: ecs.ContainerImage.fromAsset("../frontend"),
          containerPort: 3000,
        },
        memoryLimitMiB: 512,
        cpu: 256,
        publicLoadBalancer: true,
      }
    );

    // Backend Service
    const backend = new ecsPatterns.ApplicationLoadBalancedFargateService(
      this,
      "Backend",
      {
        cluster,
        taskImageOptions: {
          image: ecs.ContainerImage.fromAsset("../backend"),
          containerPort: 8080,
          environment: {
            DB_HOST: database.instanceEndpoint.hostname,
            DB_NAME: "factoryflow",
            DB_USER: "factoryuser",
          },
          secrets: {
            DB_PASSWORD: ecs.Secret.fromSecretsManager(
              database.secret!,
              "password"
            ),
          },
        },
        memoryLimitMiB: 512,
        cpu: 256,
      }
    );

    // Allow backend to access database
    database.connections.allowFrom(backend.service, ec2.Port.tcp(5432));

    // Output URLs
    new cdk.CfnOutput(this, "FrontendURL", {
      value: frontend.loadBalancer.loadBalancerDnsName,
    });

    new cdk.CfnOutput(this, "BackendURL", {
      value: backend.loadBalancer.loadBalancerDnsName,
    });
  }
}

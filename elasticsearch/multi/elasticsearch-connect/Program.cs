using Elastic.Clients.Elasticsearch;
using Elastic.Transport;

var nodes = new Uri[]
{
    new Uri("http://localhost:9201"),
    new Uri("http://localhost:9202"),
    new Uri("http://localhost:9203")
};

var pool = new StaticNodePool(nodes);

var settings = new ElasticsearchClientSettings(pool);
// .CertificateFingerprint("<FINGERPRINT>")
// .Authentication(new ApiKey("<API_KEY>"));

var client = new ElasticsearchClient(settings);
var info = await client.InfoAsync();

var response = await client.Cluster.HealthAsync<StringResponse>();
Console.WriteLine(response.Status);

System.Console.WriteLine(info.ClusterName + " " + info.Version.Number);
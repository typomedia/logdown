<?php

namespace AppBundle\Controller;

use AppBundle\Repository\LogRepository;
use AppBundle\Service\LogParser;
use PDO;
use Symfony\Bundle\FrameworkBundle\Controller\AbstractController;
use Symfony\Component\DependencyInjection\ParameterBag\ParameterBagInterface;
use Symfony\Component\Filesystem\Filesystem;
use Symfony\Component\HttpFoundation\JsonResponse;
use Symfony\Component\HttpFoundation\Request;
use Symfony\Component\HttpFoundation\Response;
use Symfony\Component\Routing\Annotation\Route;

/**
 * @Route("/upload")
 */
class UploadController extends AbstractController
{
    /**
     * @var array
     */
    public $logs;

    protected $db;
    protected $fs;

    protected $totalLines;
    /**
     * @var LogRepository
     */
    private $repo;
    /**
     * @var ParameterBagInterface
     */
    private $params;

    public function __construct(LogRepository $repo, ParameterBagInterface $params)
    {
        $this->repo = $repo;
        $this->params = $params;
        $this->fs = new Filesystem();
    }

    /**
     * @Route("/")
     */
    public function index(): Response
    {
        return $this->render('@App/upload/upload.html.twig');
    }

    /**
     * @Route("/upload")
     */
    public function upload(Request $request)
    {
        $files = $request->files->get('file');

        foreach ($files as $file) {
            $name = $file->getClientOriginalName();
            $type = $file->getMimetype();

            // only plain text logfiles
            if (in_array($type, ['text/plain'])) {
                $fs = new Filesystem();
                $fs->copy($file, $name);
            }
        }

        $schema = $this->repo->getContent('Create.sql');

        // Build the database in memory so the import never touches the disk,
        // then move the finished database to its destination in one pass.
        $this->db = new PDO('sqlite::memory:');
        $this->db->exec($schema);
        $this->db->exec('PRAGMA foreign_keys = OFF');

        $parser = new LogParser();

        $this->db->beginTransaction();
        foreach ($files as $file) {
            $name = $file->getClientOriginalName();
            $type = $file->getMimetype();

            if ($type === 'text/plain') {
                $count = 0;
                foreach ($parser->parseFile($file->getPathname()) as $row) {
                    $this->insertRow($row);
                    $count++;
                }
                $this->totalLines = $count;

                $this->logs['files'][] = $name;
            }
        }
        $this->db->commit();

        $temp = $this->params->get('kernel.project_dir') . '/var/data/sqlog.db';
        $this->persist($temp, $schema);

        return new JsonResponse($this->logs);
    }

    /**
     * Move the in-memory database to the on-disk file that Doctrine reads,
     * replacing any previous contents.
     *
     * @param string $path   Destination SQLite file.
     * @param string $schema  CREATE statement(s) for the Log table.
     * @return void
     */
    private function persist(string $path, string $schema): void
    {
        $this->fs->remove($path);
        $this->fs->mkdir(dirname($path));

        $this->db->exec(sprintf('ATTACH DATABASE %s AS disk', $this->db->quote($path)));
        $this->db->exec('BEGIN');
        $this->db->exec(preg_replace('/\bLog\b/', 'disk.Log', $schema, 1));
        $this->db->exec('INSERT INTO disk.Log SELECT * FROM main.Log');
        $this->db->exec('COMMIT');
        $this->db->exec('DETACH DATABASE disk');
    }

    /**
     * Insert a parsed log row, building the statement from the columns the
     * parser actually produced so that varying field sets are supported.
     *
     * @param array $row
     * @return void
     */
    private function insertRow(array $row): void
    {
        $columns = array_keys($row);
        $placeholders = array_map(static function (string $column) {
            return ':' . $column;
        }, $columns);

        $query = sprintf(
            'INSERT INTO Log (%s) VALUES (%s)',
            implode(', ', $columns),
            implode(', ', $placeholders)
        );

        $stmt = $this->db->prepare($query);
        foreach ($row as $column => $value) {
            $stmt->bindValue(':' . $column, $value);
        }
        $stmt->execute();
    }
}
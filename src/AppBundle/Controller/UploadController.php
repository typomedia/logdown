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

        $temp = $this->params->get('kernel.project_dir') . '/var/data/sqlog.db';
        $this->fs->remove($temp);
        $this->db = new PDO('sqlite:' . $temp);

        $query = $this->repo->getContent('Create.sql');

        $this->db->exec($query);
        $this->db->exec('PRAGMA foreign_keys = OFF');

        $parser = new LogParser();

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

        return new JsonResponse($this->logs);
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
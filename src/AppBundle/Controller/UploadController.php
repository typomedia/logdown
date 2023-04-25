<?php

namespace AppBundle\Controller;

use AppBundle\Repository\LogRepository;
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

        foreach ($files as $file) {
            $name = $file->getClientOriginalName();
            $type = $file->getMimetype();

            if ($type === 'text/plain') {
                $lines = file($file->getPathname());

                foreach ($lines as $key => $line) {
                    $entry = explode(' ', $line);

                    if ($entry[0][0] !== '#') {
                        $datetime = $entry[0] . ' ' . $entry[1];
                        $query = "INSERT INTO Log (id, date, server, method, request, param, port, user, client, agent, referer, status, substatus, win32, duration) 
VALUES (null, '$datetime', '$entry[2]', '$entry[3]', '$entry[4]', '$entry[5]', '$entry[6]', '$entry[7]', '$entry[8]', '$entry[9]', '$entry[10]', '$entry[11]', '$entry[12]', '$entry[13]', '$entry[14]')";
                        $this->db->exec($query);
                    }
                    $this->totalLines = $key + 1;
                }

                $this->logs['files'][] = $name;
            }
        }

        return new JsonResponse($this->logs);
    }
}